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

//Login logs you in to the qbittorrent client
//returns the current authentication status
func (client *Client) Login(username string, password string) (loggedIn bool, err error) {
	credentials := make(map[string]string)
	credentials["username"] = username
	credentials["password"] = password

	resp, err := client.post("login", credentials)
	if err != nil {
		return false, err
	} else if resp.Status != "200 OK" { // check for correct status code
		return false, wrapper.Wrap(ErrBadResponse, "couldnt log in")
	}

	// change authentication status so we know were authenticated in later requests
	client.Authenticated = true

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

	return client.Authenticated, nil
}

//Logout logs you out of the qbittorrent client
//returns the current authentication status
func (client *Client) Logout() (loggedOut bool, err error) {
	resp, err := client.get("logout", nil)
	if err != nil {
		return false, err
	}

	// check for correct status code
	if resp.Status != "200 OK" {
		return false, wrapper.Wrap(ErrBadResponse, "couldnt log in")
	}

	// change authentication status so we know were not authenticated in later requests
	client.Authenticated = false

	return client.Authenticated, nil
}

//Shutdown shuts down the qbittorrent client
func (client *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := client.get("command/shutdown", nil)

	// return true if successful
	return (resp.Status == "200 OK"), err
}

//Torrents returns a list of all torrents in qbittorrent matching your filter
func (client *Client) Torrents(filters map[string]string) (torrentList []BasicTorrent, err error) {
	var t []BasicTorrent
	resp, err := client.get("query/torrents", filters)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//Torrent returns a specific torrent matching the infoHash
func (client *Client) Torrent(infoHash string) (Torrent, error) {
	var t Torrent
	resp, err := client.get("query/propertiesGeneral/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentTrackers returns all trackers for a specific torrent matching the infoHash
func (client *Client) TorrentTrackers(infoHash string) ([]Tracker, error) {
	var t []Tracker
	resp, err := client.get("query/propertiesTrackers/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentWebSeeds returns seeders for a specific torrent matching the infoHash
func (client *Client) TorrentWebSeeds(infoHash string) ([]WebSeed, error) {
	var w []WebSeed
	resp, err := client.get("query/propertiesWebSeeds/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return w, err
	}
	json.NewDecoder(resp.Body).Decode(&w)
	return w, nil
}

//TorrentFiles gets the files of a specifc torrent matching the infoHash
func (client *Client) TorrentFiles(infoHash string) ([]TorrentFile, error) {
	var t []TorrentFile
	resp, err := client.get("query/propertiesFiles"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

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

//DownloadFromLink starts downloading a torrent from a link
func (client *Client) DownloadFromLink(link string, options map[string]string) (*http.Response, error) {
	options["urls"] = link
	return client.postMultipartData("command/download", options)
}

//DownloadFromFile starts downloading a torrent from a file
func (client *Client) DownloadFromFile(file string, options map[string]string) (*http.Response, error) {
	return client.postMultipartFile("command/upload", file, options)
}

//AddTrackers adds trackers to a specific torrent matching infoHash
func (client *Client) AddTrackers(infoHash string, trackers string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	params["urls"] = trackers

	return client.post("command/addTrackers", params)
}

//processInfoHashList puts list into a combined (single element) map with all hashes connected with '|'
//this is how the WEBUI API recognizes multiple hashes
func (Client) processInfoHashList(infoHashList []string) (hashMap map[string]string) {
	params := map[string]string{}
	infoHash := ""
	for i, v := range infoHashList {
		if i > 0 {
			infoHash += "|" + v
		} else {
			infoHash = v
		}
	}
	params["hashes"] = infoHash
	return params
}

//Pause pauses a specific torrent matching infoHash
func (client *Client) Pause(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)

	return client.post("command/pause", params)
}

//PauseAll pauses all torrents
func (client *Client) PauseAll() (*http.Response, error) {
	return client.get("command/pauseAll", nil)
}

//PauseMultiple pauses a list of torrents matching the infoHashes
func (client *Client) PauseMultiple(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/pauseAll", params)
}

//SetLabel sets the labels for a list of torrents matching infoHashes
func (client *Client) SetLabel(infoHashList []string, label string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	params["label"] = label

	return client.post("command/setLabel", params)
}

//SetCategory sets the category for a list of torrents matching infoHashes
func (client *Client) SetCategory(infoHashList []string, category string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	params["category"] = category

	return client.post("command/setLabel", params)
}

//Resume resumes a specific torrent matching infoHash
func (client *Client) Resume(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	return client.post("command/resume", params)
}

//ResumeAll resumes all torrents matching infoHashes
func (client *Client) ResumeAll(infoHashList []string) (*http.Response, error) {
	return client.get("command/resumeAll", nil)
}

//ResumeMultiple resumes a list of torrents matching infoHashes
func (client *Client) ResumeMultiple(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/resumeAll", params)
}

//DeleteTemp deletes the temporary files for a list of torrents matching infoHashes
func (client *Client) DeleteTemp(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/delete", params)
}

//DeletePermanently deletes all files for a list of torrents matching infoHashes
func (client *Client) DeletePermanently(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/deletePerm", params)
}

//Recheck rechecks a list of torrents matching infoHashes
func (client *Client) Recheck(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/recheck", params)
}

//IncreasePriority increases the priority of a list of torrents matching infoHashes
func (client *Client) IncreasePriority(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/increasePrio", params)
}

//DecreasePriority decreases the priority of a list of torrents matching infoHashes
func (client *Client) DecreasePriority(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/decreasePrio", params)
}

//SetMaxPriority sets the max priority for a list of torrents matching infoHashes
func (client *Client) SetMaxPriority(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/topPrio", params)
}

//SetMinPriority sets the min priority for a list of torrents matching infoHashes
func (client *Client) SetMinPriority(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/bottomPrio", params)
}

//SetFilePriority sets the priority for a specific torrent filematching infoHash
func (client *Client) SetFilePriority(infoHash string, fileID string, priority string) (*http.Response, error) {
	// disallow certain priorities that are not allowed by the WEBUI API
	priorities := [...]string{"0", "1", "2", "7"}
	for _, v := range priorities {
		if v == priority {
			return nil, ErrBadPriority
		}
	}

	params := make(map[string]string)
	params["hash"] = infoHash
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
func (client *Client) GetTorrentDownloadLimit(infoHashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := client.processInfoHashList(infoHashList)
	resp, err := client.post("command/getTorrentsDlLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentDownloadLimit sets the download limit for a list of torrents
func (client *Client) SetTorrentDownloadLimit(infoHashList []string, limit string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	params["limit"] = limit
	return client.post("command/setTorrentsDlLimit", params)
}

//GetTorrentUploadLimit gets the upload limit for a list of torrents
func (client *Client) GetTorrentUploadLimit(infoHashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := client.processInfoHashList(infoHashList)
	resp, err := client.post("command/getTorrentsUpLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentUploadLimit sets the upload limit of a list of torrents
func (client *Client) SetTorrentUploadLimit(infoHashList []string, limit string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
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
func (client *Client) ToggleSequentialDownload(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/toggleSequentialDownload", params)
}

//ToggleFirstLastPiecePriority toggles first last piece priority of a list of torrents
func (client *Client) ToggleFirstLastPiecePriority(infoHashList []string) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	return client.post("command/toggleFirstLastPiecePrio", params)
}

//ForceStart force starts a list of torrents
func (client *Client) ForceStart(infoHashList []string, value bool) (*http.Response, error) {
	params := client.processInfoHashList(infoHashList)
	params["value"] = strconv.FormatBool(value)
	return client.post("command/setForceStart", params)
}
