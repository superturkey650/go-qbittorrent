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
	//"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	wrapper "github.com/pkg/errors"

	"golang.org/x/net/publicsuffix"
)

//BadPriority means the priority is not allowd by qbittorrent
var ErrBadPriority error = errors.New("priority not available")

//BadResponse means there qbittorrent sent back an unexpected response
var ErrBadResponse error = errors.New("received bad response")

//Client creates a connection to qbittorrent and performs requests
type Client struct {
	http          *http.Client
	URL           string
	Authenticated bool
	Jar           http.CookieJar
}

//NewClient creates a new client connection to qbittorrent
func NewClient(url string) *Client {
	c := &Client{}

	// ensure url ends with "/"
	if url[len(url)-1:] != "/" {
		url = url + "/"
	}

	c.URL = url

	// create cookie jar
	c.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c.http = &http.Client{
		Jar: c.Jar,
	}
	return c
}

//get will perform a GET request with no parameters
func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL+endpoint, nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to build request")
	}

	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		q := req.URL.Query()
		for k, v := range opts {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//post will perform a POST request with no content-type specified
func (c *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.URL+endpoint, nil)
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

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil

}

//postMultipart will perform a multiple part POST request
func (c *Client) postMultipart(endpoint string, b bytes.Buffer, cType string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.URL+endpoint, &b)
	if err != nil {
		return nil, wrapper.Wrap(err, "error creating request")
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", cType)
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//writeOptions will write a map to the buffer through multipart.NewWriter
func writeOptions(w *multipart.Writer, opts map[string]string) {
	for key, val := range opts {
		w.WriteField(key, val)
	}
}

//postMultipartData will perform a multiple part POST request without a file
func (c *Client) postMultipartData(endpoint string, opts map[string]string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// write the options to the buffer
	// will contain the link string
	writeOptions(w, opts)

	// close the writer before doing request to get closing line on multipart request
	if err := w.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := c.postMultipart(endpoint, b, w.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//postMultipartFile will perform a multiple part POST request with a file
func (c *Client) postMultipartFile(endpoint string, file string, opts map[string]string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// open the file for reading
	f, err := os.Open(file)
	if err != nil {
		return nil, wrapper.Wrap(err, "error opening file")
	}
	// defer the closing of the file until the end of function
	// so we can still copy its contents
	defer f.Close()

	// create form for writing the file to and give it the filename
	fw, err := w.CreateFormFile("torrents", path.Base(file))
	if err != nil {
		return nil, wrapper.Wrap(err, "error adding file")
	}

	// write the options to the buffer
	writeOptions(w, opts)

	// copy the file contents into the form
	if _, err = io.Copy(fw, f); err != nil {
		return nil, wrapper.Wrap(err, "error copying file")
	}

	// close the writer before doing request to get closing line on multipart request
	if err := w.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := c.postMultipart(endpoint, b, w.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//Login logs you in to the qbittorrent client
//returns the current authentication status
func (c *Client) Login(username string, password string) (loggedIn bool, err error) {
	creds := make(map[string]string)
	creds["username"] = username
	creds["password"] = password

	resp, err := c.post("login", creds)
	if err != nil {
		return false, err
	}

	// check for correct status code
	if resp.Status != "200 OK" {
		err = ErrBadResponse
		return false, wrapper.Wrap(err, "couldnt log in")
	}

	// change authentication status so we know were authenticated in later requests
	c.Authenticated = true

	// add the cookie to cookie jar to authenticate later requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		cookieURL, _ := url.Parse("http://localhost:8080")
		c.Jar.SetCookies(cookieURL, cookies)
	}

	// create a new client with the cookie jar and replace the old one
	// so that all our later requests are authenticated
	c.http = &http.Client{
		Jar: c.Jar,
	}

	return c.Authenticated, nil
}

//Logout logs you out of the qbittorrent client
//returns the current authentication status
func (c *Client) Logout() (loggedOut bool, err error) {
	resp, err := c.get("logout", nil)
	if err != nil {
		return false, err
	}

	// check for correct status code
	if resp.Status != "200 OK" {
		err = ErrBadResponse
		return false, wrapper.Wrap(err, "couldnt log in")
	}

	// change authentication status so we know were not authenticated in later requests
	c.Authenticated = false

	return c.Authenticated, nil
}

//Shutdown shuts down the qbittorrent client
func (c *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := c.get("command/shutdown", nil)

	// return true if successful
	return resp.Status == "200 OK", err
}

//Torrents returns a list of all torrents in qbittorrent matching your filter
func (c *Client) Torrents(filters map[string]string) (torrentList []BasicTorrent, err error) {
	var t []BasicTorrent
	resp, err := c.get("query/torrents", filters)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//Torrent returns a specific torrent matching the infoHash
func (c *Client) Torrent(infoHash string) (Torrent, error) {
	var t Torrent
	resp, err := c.get("query/propertiesGeneral/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentTrackers returns all trackers for a specific torrent matching the infoHash
func (c *Client) TorrentTrackers(infoHash string) ([]Tracker, error) {
	var t []Tracker
	resp, err := c.get("query/propertiesTrackers/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentWebSeeds returns seeders for a specific torrent matching the infoHash
func (c *Client) TorrentWebSeeds(infoHash string) ([]WebSeed, error) {
	var w []WebSeed
	resp, err := c.get("query/propertiesWebSeeds/"+strings.ToLower(infoHash), nil)
	if err != nil {
		return w, err
	}
	json.NewDecoder(resp.Body).Decode(&w)
	return w, nil
}

//TorrentFiles gets the files of a specifc torrent matching the infoHash
func (c *Client) TorrentFiles(infoHash string) ([]TorrentFile, error) {
	var t []TorrentFile
	resp, err := c.get("query/propertiesFiles"+strings.ToLower(infoHash), nil)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//Sync returns the server state and list of torrents in one struct
func (c *Client) Sync(rid string) (Sync, error) {
	var s Sync

	params := make(map[string]string)
	params["rid"] = rid

	resp, err := c.get("sync/maindata", params)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//DownloadFromLink starts downloading a torrent from a link
func (c *Client) DownloadFromLink(link string, options map[string]string) (*http.Response, error) {
	options["urls"] = link
	return c.postMultipartData("command/download", options)
}

//DownloadFromFile starts downloading a torrent from a file
func (c *Client) DownloadFromFile(file string, options map[string]string) (*http.Response, error) {
	return c.postMultipartFile("command/upload", file, options)
}

//AddTrackers adds trackers to a specific torrent matching infoHash
func (c *Client) AddTrackers(infoHash string, trackers string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	params["urls"] = trackers

	return c.post("command/addTrackers", params)
}

//processInfoHashList puts list into a combined (single element) map with all hashes connected with '|'
//this is how the WEBUI API recognizes multiple hashes
func (Client) processInfoHashList(infoHashList []string) (hashMap map[string]string) {
	d := map[string]string{}
	infoHash := ""
	for i, v := range infoHashList {
		if i > 0 {
			infoHash += "|" + v
		} else {
			infoHash = v
		}
	}
	d["hashes"] = infoHash
	return d
}

//Pause pauses a specific torrent matching infoHash
func (c *Client) Pause(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)

	return c.post("command/pause", params)
}

//PauseAll pauses all torrents
func (c *Client) PauseAll() (*http.Response, error) {
	return c.get("command/pauseAll", nil)
}

//PauseMultiple pauses a list of torrents matching the infoHashes
func (c *Client) PauseMultiple(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/pauseAll", params)
}

//SetLabel sets the labels for a list of torrents matching infoHashes
func (c *Client) SetLabel(infoHashList []string, label string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["label"] = label

	return c.post("command/setLabel", params)
}

//SetCategory sets the category for a list of torrents matching infoHashes
func (c *Client) SetCategory(infoHashList []string, category string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["category"] = category

	return c.post("command/setLabel", params)
}

//Resume resumes a specific torrent matching infoHash
func (c *Client) Resume(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	return c.post("command/resume", params)
}

//ResumeAll resumes all torrents matching infoHashes
func (c *Client) ResumeAll(infoHashList []string) (*http.Response, error) {
	return c.get("command/resumeAll", nil)
}

//ResumeMultiple resumes a list of torrents matching infoHashes
func (c *Client) ResumeMultiple(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/resumeAll", params)
}

//DeleteTemp deletes the temporary files for a list of torrents matching infoHashes
func (c *Client) DeleteTemp(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/delete", params)
}

//DeletePermanently deletes all files for a list of torrents matching infoHashes
func (c *Client) DeletePermanently(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/deletePerm", params)
}

//Recheck rechecks a list of torrents matching infoHashes
func (c *Client) Recheck(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/recheck", params)
}

//IncreasePriority increases the priority of a list of torrents matching infoHashes
func (c *Client) IncreasePriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/increasePrio", params)
}

//DecreasePriority decreases the priority of a list of torrents matching infoHashes
func (c *Client) DecreasePriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/decreasePrio", params)
}

//SetMaxPriority sets the max priority for a list of torrents matching infoHashes
func (c *Client) SetMaxPriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/topPrio", params)
}

//SetMinPriority sets the min priority for a list of torrents matching infoHashes
func (c *Client) SetMinPriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/bottomPrio", params)
}

//SetFilePriority sets the priority for a specific torrent filematching infoHash
func (c *Client) SetFilePriority(infoHash string, fileID string, priority string) (*http.Response, error) {
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

	return c.post("command/setFilePriority", params)
}

//GetGlobalDownloadLimit gets the global download limit of your qbittorrent client
func (c *Client) GetGlobalDownloadLimit() (limit int, err error) {
	var l int
	resp, err := c.get("command/getGlobalDlLimit", nil)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetGlobalDownloadLimit sets the global download limit of your qbittorrent client
func (c *Client) SetGlobalDownloadLimit(limit string) (*http.Response, error) {
	params := make(map[string]string)
	params["limit"] = limit
	return c.post("command/setGlobalDlLimit", params)
}

//GetGlobalUploadLimit gets the global upload limit of your qbittorrent client
func (c *Client) GetGlobalUploadLimit() (limit int, err error) {
	var l int
	resp, err := c.get("command/getGlobalUpLimit", nil)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetGlobalUploadLimit sets the global upload limit of your qbittorrent client
func (c *Client) SetGlobalUploadLimit(limit string) (*http.Response, error) {
	params := make(map[string]string)
	params["limit"] = limit
	return c.post("command/setGlobalUpLimit", params)
}

//GetTorrentDownloadLimit gets the download limit for a list of torrents
func (c *Client) GetTorrentDownloadLimit(infoHashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := c.processInfoHashList(infoHashList)
	resp, err := c.post("command/getTorrentsDlLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentDownloadLimit sets the download limit for a list of torrents
func (c *Client) SetTorrentDownloadLimit(infoHashList []string, limit string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["limit"] = limit
	return c.post("command/setTorrentsDlLimit", params)
}

//GetTorrentUploadLimit gets the upload limit for a list of torrents
func (c *Client) GetTorrentUploadLimit(infoHashList []string) (limits map[string]string, err error) {
	var l map[string]string
	params := c.processInfoHashList(infoHashList)
	resp, err := c.post("command/getTorrentsUpLimit", params)
	if err != nil {
		return l, err
	}
	json.NewDecoder(resp.Body).Decode(&l)
	return l, nil
}

//SetTorrentUploadLimit sets the upload limit of a list of torrents
func (c *Client) SetTorrentUploadLimit(infoHashList []string, limit string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["limit"] = limit
	return c.post("command/setTorrentsUpLimit", params)
}

//SetPreferences sets the preferences of your qbittorrent client
func (c *Client) SetPreferences(params map[string]string) (*http.Response, error) {
	return c.post("command/setPreferences", params)
}

//GetAlternativeSpeedStatus gets the alternative speed status of your qbittorrent client
func (c *Client) GetAlternativeSpeedStatus() (status bool, err error) {
	var s bool
	resp, err := c.get("command/alternativeSpeedLimitsEnabled", nil)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//ToggleAlternativeSpeed toggles the alternative speed of your qbittorrent client
func (c *Client) ToggleAlternativeSpeed() (*http.Response, error) {
	return c.get("command/toggleAlternativeSpeedLimits", nil)
}

//ToggleSequentialDownload toggles the download sequence of a list of torrents
func (c *Client) ToggleSequentialDownload(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/toggleSequentialDownload", params)
}

//ToggleFirstLastPiecePriority toggles first last piece priority of a list of torrents
func (c *Client) ToggleFirstLastPiecePriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/toggleFirstLastPiecePrio", params)
}

//ForceStart force starts a list of torrents
func (c *Client) ForceStart(infoHashList []string, value bool) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["value"] = strconv.FormatBool(value)
	return c.post("command/setForceStart", params)
}
