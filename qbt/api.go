package qbt

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"

	"net/url"
	"strconv"
	"strings"

	wrapper "github.com/pkg/errors"

	"golang.org/x/net/publicsuffix"
)

//ErrBadPriority means the priority is not allowd by qbittorrent
var ErrBadPriority = errors.New("priority not available")

//ErrBadResponse means that qbittorrent sent back an unexpected response
var ErrBadResponse = errors.New("received bad response")

//Client creates a connection to qbittorrent and performs requests
type Client struct {
	http          *http.Client
	URL           string
	Authenticated bool
	Jar           http.CookieJar
}

//NewClient creates a new client connection to qbittorrent
func NewClient(url string) *Client {
	client := &Client{}

	// ensure url ends with "/"
	if url[len(url)-1:] != "/" {
		url += "/"
	}

	client.URL = url

	// create cookie jar
	client.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client.http = &http.Client{
		Jar: client.Jar,
	}
	return client
}

//get will perform a GET request with no parameters
func (client *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", client.URL+endpoint, nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to build request")
	}

	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		query := req.URL.Query()
		for k, v := range opts {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//post will perform a POST request with no content-type specified
func (client *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", client.URL+endpoint, nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to build request")
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		form := url.Values{}
		for k, v := range opts {
			form.Add(k, v)
		}
		req.PostForm = form
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil

}

//postMultipart will perform a multiple part POST request
func (client *Client) postMultipart(endpoint string, buffer bytes.Buffer, contentType string) (*http.Response, error) {
	req, err := http.NewRequest("POST", client.URL+endpoint, &buffer)
	if err != nil {
		return nil, wrapper.Wrap(err, "error creating request")
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", contentType)
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//writeOptions will write a map to the buffer through multipart.NewWriter
func writeOptions(writer *multipart.Writer, opts map[string]string) {
	for key, val := range opts {
		writer.WriteField(key, val)
	}
}

//postMultipartData will perform a multiple part POST request without a file
func (client *Client) postMultipartData(endpoint string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// write the options to the buffer
	// will contain the link string
	writeOptions(writer, opts)

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := client.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//postMultipartFile will perform a multiple part POST request with a file
func (client *Client) postMultipartFile(endpoint string, fileName string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		return nil, wrapper.Wrap(err, "error opening file")
	}
	// defer the closing of the file until the end of function
	// so we can still copy its contents
	defer file.Close()

	// create form for writing the file to and give it the filename
	formWriter, err := writer.CreateFormFile("torrents", path.Base(fileName))
	if err != nil {
		return nil, wrapper.Wrap(err, "error adding file")
	}

	// write the options to the buffer
	writeOptions(writer, opts)

	// copy the file contents into the form
	if _, err = io.Copy(formWriter, file); err != nil {
		return nil, wrapper.Wrap(err, "error copying file")
	}

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := client.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Application endpoints

//Login logs you in to the qbittorrent client
//returns the current authentication status
func (client *Client) Login(username string, password string) (loggedIn bool, err error) {
	params := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := client.post("api/v2/auth/login", credentials)
	if err != nil {
		return loggedIn, err
	} else if resp.Status != "200 OK" { // check for correct status code
		return loggedIn, wrapper.Wrap(ErrBadResponse, "couldnt log in")
	}

	// change authentication status so we know were authenticated in later requests
	loggedIn = true
	client.Authenticated = loggedIn

	// add the cookie to cookie jar to authenticate later requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		cookieURL, _ := url.Parse("http://localhost:8080")
		client.Jar.SetCookies(cookieURL, cookies)
	}

	// create a new client with the cookie jar and replace the old one
	// so that all our later requests are authenticated
	client.http = &http.Client{
		Jar: client.Jar,
	}

	return loggedIn, nil
}

//Logout logs you out of the qbittorrent client
//returns the current authentication status
func (client *Client) Logout() (loggedOut bool, err error) {
	resp, err := client.get("api/v2/auth/logout", nil)
	if err != nil {
		return loggedOut, err
	}

	// check for correct status code
	if resp.Status != "200 OK" {
		return loggedOut, wrapper.Wrap(ErrBadResponse, "couldnt log in")
	}

	// change authentication status so we know were not authenticated in later requests
	loggedOut = false
	client.Authenticated = loggedOut

	return loggedOut, nil
}

//ApplicationVersion of the qbittorrent client
func (client *Client) ApplicationVersion() (version string, err error) {
	resp, err := client.get("/api/v2/app/version", nil)
	if err != nil {
		return version, err
	}
	json.NewDecoder(resp.Body).Decode(&version)
	return version, err
}

//WebAPIVersion of the qbittorrent client
func (client *Client) WebAPIVersion() (version string, err error) {
	resp, err := client.get("/api/v2/app/webapiVersion", nil)
	if err != nil {
		return version, err
	}
	json.NewDecoder(resp.Body).Decode(&version)
	return version, err
}

//BuildInfo of the qbittorrent client
func (client *Client) BuildInfo() (buildInfo BuildInfo, err error) {
	resp, err := client.get("/api/v2/app/buildInfo", nil)
	if err != nil {
		return buildInfo, err
	}
	json.NewDecoder(resp.Body).Decode(&buildInfo)
	return buildInfo, err
}

//Preferences of the qbittorrent client
func (client *Client) Preferences() (prefs Preferences, err error) {
	resp, err := client.get("/api/v2/app/preferences", nil)
	if err != nil {
		return prefs, err
	}
	json.NewDecoder(resp.Body).Decode(&prefs)
	return prefs, err
}

//SetPreferences of the qbittorrent client
func (client *Client) SetPreferences() (prefsSet bool, err error) {
	resp, err := client.post("/api/v2/app/setPreferences", nil)
	return (resp.Status == "200 OK"), err
}

//DefaultSavePath of the qbittorrent client
func (client *Client) DefaultSavePath() (path string, err error) {
	resp, err := client.get("/api/v2/app/defaultSavePath", nil)
	if err != nil {
		return path, err
	}
	json.NewDecoder(resp.Body).Decode(&path)
	return path, err
}

//Shutdown shuts down the qbittorrent client
func (client *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := client.get("/api/v2/app/shutdown", nil)

	// return true if successful
	return (resp.Status == "200 OK"), err
}

// Log Endpoints

//Logs of the qbittorrent client
func (client *Client) Logs(filters map[string]string) (logs []Log, err error) {
	resp, err := client.get("/api/v2/log/main", filters)
	if err != nil {
		return logs, err
	}
	json.NewDecoder(resp.Body).Decode(&logs)
	return logs, err
}

//PeerLogs of the qbittorrent client
func (client *Client) PeerLogs(filters map[string]string) (logs []PeerLog, err error) {
	resp, err := client.get("/api/v2/log/peers", filters)
	if err != nil {
		return logs, err
	}
	json.NewDecoder(resp.Body).Decode(&logs)
	return logs, err
}

// TODO: Sync Endpoints

// TODO: Transfer Endpoints

//Info returns info you usually see in qBt status bar.
func (client *Client) Info() (info Info, err error) {
	resp, err := client.get("/api/v2/transfer/info", nil)
	if err != nil {
		return info, err
	}
	json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

//AltSpeedLimitsEnabled returns info you usually see in qBt status bar.
func (client *Client) AltSpeedLimitsEnabled() (mode bool, err error) {
	resp, err := client.get("/api/v2/transfer/speedLimitsMode", nil)
	if err != nil {
		return mode, err
	}
	var decoded int
	json.NewDecoder(resp.Body).Decode(&decoded)
	mode = decoded == 1
	return mode, err
}

//ToggleAltSpeedLimits returns info you usually see in qBt status bar.
func (client *Client) ToggleAltSpeedLimits() (toggled bool, err error) {
	resp, err := client.get("/api/v2/transfer/toggleSpeedLimitsMode", nil)
	if err != nil {
		return toggled, err
	}
	return (resp.Status == "200 OK"), err
}

//DlLimit returns info you usually see in qBt status bar.
func (client *Client) DlLimit() (dlLimit int, err error) {
	resp, err := client.get("/api/v2/transfer/downloadLimit", nil)
	if err != nil {
		return dlLimit, err
	}
	json.NewDecoder(resp.Body).Decode(&dlLimit)
	return dlLimit, err
}

//SetDlLimit returns info you usually see in qBt status bar.
func (client *Client) SetDlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := client.get("/api/v2/transfer/setDownloadLimit", params)
	if err != nil {
		return set, err
	}
	return (resp.Status == "200 OK"), err
}

//UlLimit returns info you usually see in qBt status bar.
func (client *Client) UlLimit() (ulLimit int, err error) {
	resp, err := client.get("/api/v2/transfer/uploadLimit", nil)
	if err != nil {
		return ulLimit, err
	}
	json.NewDecoder(resp.Body).Decode(&ulLimit)
	return ulLimit, err
}

//SetUlLimit returns info you usually see in qBt status bar.
func (client *Client) SetUlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := client.get("/api/v2/transfer/setUploadLimit", params)
	if err != nil {
		return set, err
	}
	return (resp.Status == "200 OK"), err
}

//Torrents returns a list of all torrents in qbittorrent matching your filter
func (client *Client) Torrents(filters map[string]string) (torrentList []BasicTorrent, err error) {
	resp, err := client.get("/api/v2/torrents/info", filters)
	if err != nil {
		return torrentList, err
	}
	json.NewDecoder(resp.Body).Decode(&torrentList)
	return torrentList, nil
}

//Torrent returns a specific torrent matching the hash
func (client *Client) Torrent(hash string) (torrent Torrent, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/properties", opts)
	if err != nil {
		return torrent, err
	}
	json.NewDecoder(resp.Body).Decode(&torrent)
	return torrent, nil
}

//TorrentTrackers returns all trackers for a specific torrent matching the hash
func (client *Client) TorrentTrackers(hash string) (trackers []Tracker, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/trackers", opts)
	if err != nil {
		return trackers, err
	}
	json.NewDecoder(resp.Body).Decode(&trackers)
	return trackers, nil
}

//TorrentWebSeeds returns seeders for a specific torrent matching the hash
func (client *Client) TorrentWebSeeds(hash string) (webSeeds []WebSeed, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/webseeds", opts)
	if err != nil {
		return webSeeds, err
	}
	json.NewDecoder(resp.Body).Decode(&webSeeds)
	return webSeeds, nil
}

//TorrentFiles
func (client *Client) TorrentFiles(hash string) (files []TorrentFile, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/files", opts)
	if err != nil {
		return files, err
	}
	json.NewDecoder(resp.Body).Decode(&files)
	return files, nil
}

//TorrentPieceStates
func (client *Client) TorrentPieceStates(hash string) (states []int, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/pieceStates", opts)
	if err != nil {
		return states, err
	}
	json.NewDecoder(resp.Body).Decode(&states)
	return states, nil
}

//TorrentPieceHashes
func (client *Client) TorrentPieceHashes(hash string) (hashes []string, err error) {
	var opts = map[string]string{"hash", strings.ToLower(hash)}
	resp, err := client.get("/api/v2/torrents/pieceHashes", opts)
	if err != nil {
		return hashes, err
	}
	json.NewDecoder(resp.Body).Decode(&hashes)
	return hashes, nil
}

//Pause torrents
func (client *Client) Pause(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/pause", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//Resume torrents
func (client *Client) Resume(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/resume", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//Delete torrents and optionally delete their files
func (client *Client) Delete(hashes []string, deleteFiles bool) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	otps["deleteFiles"] = deleteFiles
	resp, err := client.get("/api/v2/torrents/delete", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//Recheck torrents
func (client *Client) Recheck(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/recheck", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//Reannounce torrents
func (client *Client) Reannounce(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/reannounce", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//DownloadFromLink starts downloading a torrent from a link
func (client *Client) DownloadFromLink(link string, options map[string]string) (*http.Response, error) {
	options["urls"] = link
	restp, err := client.postMultipartData("/api/v2/torrents/add", options)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//DownloadFromFile starts downloading a torrent from a file
func (client *Client) DownloadFromFile(file string, options map[string]string) (*http.Response, error) {
	resp, err := client.postMultipartFile("/api/v2/torrents/add", file, options)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil
}

//AddTrackers to a torrent
func (client *Client) AddTrackers(hash string, trackers string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(hash)
	params["urls"] = trackers // add escaping for ampersand in urls

	return client.post("/api/v2/torrents/addTrackers", params)
}

//EditTracker on a torrent
func (client *Client) Recheck(hash string, originalURL string, newURL string) (bool, error) {
	opts := map[string]string{
		"hash":    hash,
		"origUrl": originalURL,
		"newUrl":  newURL,
	}
	resp, err := client.get("/api/v2/torrents/editTracker", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//RemoveTrackers from a torrent
func (client *Client) RemoveTrackers(hash string, urls []string) (bool, error) {
	opts := map[string]string{
		"hash": hash,
		"urls": piper(urls),
	}
	resp, err := client.get("/api/v2/torrents/removeTrackers", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//IncreasePriority of torrents
func (client *Client) IncreasePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/IncreasePrio", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//DecreasePriority of torrents
func (client *Client) DecreasePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/DecreasePrio", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//MaxPriority maximizes the priority of torrents
func (client *Client) MaxPriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/TopPrio", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//MinPriority maximizes the priority of torrents
func (client *Client) MinPriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": piper(hashes)}
	resp, err := client.get("/api/v2/torrents/BottomPrio", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

//SetFilePriority for a torrent
func (client *Client) MinPriority(hash string, ids []string, priority int) (bool, error) {
	opts := map[string]string{
		"hashes":   hash,
		"id":       piper(ids),
		"priority": strcnv.itoa(priority),
	}
	resp, err := client.get("/api/v2/torrents/filePrio", opts)
	if err != nil {
		return nil, err
	}

	return resp.StatusCode == "200 OK", nil //TODO: look into other statuses
}

// New Above
//
//
//
//
//
//
//
//
//
//
//
//
// Old Below

//Sync returns the server state and list of torrents in one struct
func (client *Client) Sync(rid string) (Sync, error) {
	var s Sync

	params := make(map[string]string)
	params["rid"] = rid

	resp, err := client.get("sync/maindata", params)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//piper puts list into a combined (single element) map with all hashes connected with '|'
//this is how the WEBUI API recognizes multiple hashes
func piper(items []string) (piped string) {
	for i, v := range items {
		if i > 0 {
			piped += "|" + v
		} else {
			piped = v
		}
	}
	return piped
}

//SetLabel sets the labels for a list of torrents matching hashes
func (client *Client) SetLabel(hashList []string, label string) (*http.Response, error) {
	params := client.processhashList(hashList)
	params["label"] = label

	return client.post("command/setLabel", params)
}

//SetCategory sets the category for a list of torrents matching hashes
func (client *Client) SetCategory(hashList []string, category string) (*http.Response, error) {
	params := client.processhashList(hashList)
	params["category"] = category

	return client.post("command/setLabel", params)
}

//SetFilePriority sets the priority for a specific torrent filematching hash
func (client *Client) SetFilePriority(hash string, fileID string, priority string) (*http.Response, error) {
	// disallow certain priorities that are not allowed by the WEBUI API
	priorities := [...]string{"0", "1", "2", "7"}
	for _, v := range priorities {
		if v == priority {
			return nil, ErrBadPriority
		}
	}

	params := make(map[string]string)
	params["hash"] = hash
	params["id"] = fileID
	params["priority"] = priority

	return client.post("command/setFilePriority", params)
}

//GetGlobalDownloadLimit gets the global download limit of your qbittorrent client
func (client *Client) GetGlobalDownloadLimit() (limit int, err error) {
	var l int
	resp, err := client.get("command/getGlobalDlLimit", nil)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetGlobalDownloadLimit sets the global download limit of your qbittorrent client
func (client *Client) SetGlobalDownloadLimit(limit string) (*http.Response, error) {
	params := make(map[string]string)
	params["limit"] = limit
	return client.post("command/setGlobalDlLimit", params)
}

//GetGlobalUploadLimit gets the global upload limit of your qbittorrent client
func (client *Client) GetGlobalUploadLimit() (limit int, err error) {
	var l int
	resp, err := client.get("command/getGlobalUpLimit", nil)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetGlobalUploadLimit sets the global upload limit of your qbittorrent client
func (client *Client) SetGlobalUploadLimit(limit string) (*http.Response, error) {
	params := make(map[string]string)
	params["limit"] = limit
	return client.post("command/setGlobalUpLimit", params)
}

//GetTorrentDownloadLimit gets the download limit for a list of torrents
func (client *Client) GetTorrentDownloadLimit(hashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := client.processhashList(hashList)
	resp, err := client.post("command/getTorrentsDlLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentDownloadLimit sets the download limit for a list of torrents
func (client *Client) SetTorrentDownloadLimit(hashList []string, limit string) (*http.Response, error) {
	params := client.processhashList(hashList)
	params["limit"] = limit
	return client.post("command/setTorrentsDlLimit", params)
}

//GetTorrentUploadLimit gets the upload limit for a list of torrents
func (client *Client) GetTorrentUploadLimit(hashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := client.processhashList(hashList)
	resp, err := client.post("command/getTorrentsUpLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentUploadLimit sets the upload limit of a list of torrents
func (client *Client) SetTorrentUploadLimit(hashList []string, limit string) (*http.Response, error) {
	params := client.processhashList(hashList)
	params["limit"] = limit
	return client.post("command/setTorrentsUpLimit", params)
}

//SetPreferences sets the preferences of your qbittorrent client
func (client *Client) SetPreferences(params map[string]string) (*http.Response, error) {
	return client.post("command/setPreferences", params)
}

//GetAlternativeSpeedStatus gets the alternative speed status of your qbittorrent client
func (client *Client) GetAlternativeSpeedStatus() (status bool, err error) {
	var s bool
	resp, err := client.get("command/alternativeSpeedLimitsEnabled", nil)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//ToggleAlternativeSpeed toggles the alternative speed of your qbittorrent client
func (client *Client) ToggleAlternativeSpeed() (*http.Response, error) {
	return client.get("command/toggleAlternativeSpeedLimits", nil)
}

//ToggleSequentialDownload toggles the download sequence of a list of torrents
func (client *Client) ToggleSequentialDownload(hashList []string) (*http.Response, error) {
	params := client.processhashList(hashList)
	return client.post("command/toggleSequentialDownload", params)
}

//ToggleFirstLastPiecePriority toggles first last piece priority of a list of torrents
func (client *Client) ToggleFirstLastPiecePriority(hashList []string) (*http.Response, error) {
	params := client.processhashList(hashList)
	return client.post("command/toggleFirstLastPiecePrio", params)
}

//ForceStart force starts a list of torrents
func (client *Client) ForceStart(hashList []string, value bool) (*http.Response, error) {
	params := client.processhashList(hashList)
	params["value"] = strconv.FormatBool(value)
	return client.post("command/setForceStart", params)
}
