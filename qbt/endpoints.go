package qbit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	//"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

//Client creates a connection to qbittorrent and performs requests
type Client struct {
	http          *http.Client
	URL           string
	Authenticated bool
	Session       string //replace with session type
	Jar           http.CookieJar
}

//NewClient creates a new client connection to qbittorrent
func NewClient(url string) *Client {
	c := &Client{}

	if url[len(url)-1:] != "/" {
		url = url + "/"
	}

	c.URL = url

	c.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c.http = &http.Client{
		Jar: c.Jar,
	}
	return c
}

//Login logs you in to the qbittorrent client
func (c *Client) Login(username string, password string) (loggedIn bool, err error) {
	creds := make(map[string]string)
	creds["username"] = username
	creds["password"] = password

	resp, err := c.post("login", creds)
	if err != nil {
		return false, err
	}
	cookieURL, _ := url.Parse("http://localhost:8080")

	if cookies := resp.Cookies(); len(cookies) > 0 {
		c.Jar.SetCookies(cookieURL, cookies)
	}

	c.http = &http.Client{
		Jar: c.Jar,
	}

	if resp.Status == "200 OK" {
		c.Authenticated = true
	} else {
		err = errors.New("received bad response")
		return false, errors.Wrap(err, "couldnt log in")
	}
	return c.Authenticated, nil
}

//Logout logs you out of the qbittorrent client
func (c *Client) Logout() (loggedOut bool, err error) {
	resp, err := c.get("logout")
	if err != nil {
		return false, err
	}
	fmt.Println(resp)
	if resp.Status == "200 OK" {
		c.Authenticated = false
	} else {
		err = errors.New("recieved bad response")
		return false, errors.Wrap(err, "couldn't log out")
	}
	return c.Authenticated, nil
}

//Shutdown shuts down the qbittorrent client
func (c *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := c.get("command/shutdown")
	return resp.Status == "200 OK", err
}

//Torrents gets a list of all torrents in qbittorrent matching your filter
func (c *Client) Torrents(filters map[string]string) (torrentList []BasicTorrent, err error) {
	var t []BasicTorrent
	params := make(map[string]string)
	for k, v := range filters {
		if k == "status" {
			k = "filter"
		}
		params[k] = v
	}
	resp, err := c.getWithParams("query/torrents", params)
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//Torrent gets a specific torrent
func (c *Client) Torrent(infoHash string) (Torrent, error) {
	var t Torrent
	resp, err := c.get("query/propertiesGeneral/" + strings.ToLower(infoHash))
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentTrackers gets all trackers for a specific torrent
func (c *Client) TorrentTrackers(infoHash string) ([]Tracker, error) {
	var t []Tracker
	resp, err := c.get("query/propertiesTrackers/" + strings.ToLower(infoHash))
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//TorrentWebSeeds gets seeders for a specific torrent
func (c *Client) TorrentWebSeeds(infoHash string) ([]WebSeed, error) {
	var w []WebSeed
	resp, err := c.get("query/propertiesWebSeeds/" + strings.ToLower(infoHash))
	if err != nil {
		return w, err
	}
	json.NewDecoder(resp.Body).Decode(&w)
	return w, nil
}

//TorrentFiles gets the files of a specifc torrent
func (c *Client) TorrentFiles(infoHash string) ([]TorrentFile, error) {
	var t []TorrentFile
	resp, err := c.get("query/propertiesFiles" + strings.ToLower(infoHash))
	if err != nil {
		return t, err
	}
	json.NewDecoder(resp.Body).Decode(&t)
	return t, nil
}

//Sync syncs main data of qbittorrent
func (c *Client) Sync(rid string) (Sync, error) {
	var s Sync
	params := make(map[string]string)
	params["rid"] = rid
	resp, err := c.getWithParams("sync/maindata", params)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//DownloadFromLink starts downloading a torrent from a link
func (c *Client) DownloadFromLink(link string, options map[string]string) (*http.Response, error) {
	options["urls"] = link
	return c.postMultipart("command/download", options)
}

//DownloadFromFile downloads a torrent from a file
func (c *Client) DownloadFromFile(file string, options map[string]string) (*http.Response, error) {
	return c.postMultipartFile("command/upload", file, options)
}

//AddTrackers adds trackers to a specific torrent
func (c *Client) AddTrackers(infoHash string, trackers string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	params["urls"] = trackers
	return c.post("command/addTrackers", params)
}

//process the hash list and put it into a combined (single element) map with all hashes connected with '|'
func (Client) processInfoHashList(infoHashList []string) (hashMap map[string]string) {
	d := map[string]string{}
	infoHash := ""
	for _, v := range infoHashList {
		infoHash = infoHash + "|" + v
	}
	d["hashes"] = infoHash
	return d
}

//Pause pauses a specific torrent
func (c *Client) Pause(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	return c.post("command/pause", params)
}

//PauseAll pauses all torrents
func (c *Client) PauseAll() (*http.Response, error) {
	return c.get("command/pauseAll")
}

//PauseMultiple pauses a list of torrents
func (c *Client) PauseMultiple(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/pauseAll", params)
}

//SetLabel sets the labels for a list of torrents
func (c *Client) SetLabel(infoHashList []string, label string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["label"] = label
	return c.post("command/setLabel", params)
}

//SetCategory sets the category for a list of torrents
func (c *Client) SetCategory(infoHashList []string, category string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	params["category"] = category
	return c.post("command/setLabel", params)
}

//Resume resumes a specific torrent
func (c *Client) Resume(infoHash string) (*http.Response, error) {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(infoHash)
	return c.post("command/resume", params)
}

//ResumeAll resumes all torrents
func (c *Client) ResumeAll(infoHashList []string) (*http.Response, error) {
	return c.get("command/resumeAll")
}

//ResumeMultiple resumes a list of torrents
func (c *Client) ResumeMultiple(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/resumeAll", params)
}

//DeleteTemp deletes the temporary files for a list of torrents
func (c *Client) DeleteTemp(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/delete", params)
}

//DeletePermanently deletes all files for a list of torrents
func (c *Client) DeletePermanently(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/deletePerm", params)
}

//Recheck rechecks a list of torrents
func (c *Client) Recheck(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/recheck", params)
}

//IncreasePriority increases the priority of a list of torrents
func (c *Client) IncreasePriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/increasePrio", params)
}

//DecreasePriority decreases the priority of a list of torrents
func (c *Client) DecreasePriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/decreasePrio", params)
}

//SetMaxPriority sets the max priority for a list of torrents
func (c *Client) SetMaxPriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/topPrio", params)
}

//SetMinPriority sets the min priority for a list of torrents
func (c *Client) SetMinPriority(infoHashList []string) (*http.Response, error) {
	params := c.processInfoHashList(infoHashList)
	return c.post("command/bottomPrio", params)
}

//SetFilePriority sets the priority for a specific torrent file
func (c *Client) SetFilePriority(infoHash string, fileID string, priority string) (*http.Response, error) {
	priorities := [...]string{"0", "1", "2", "7"}
	for _, v := range priorities {
		if v == priority {
			fmt.Println("error, priority no tavailable")
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
	resp, err := c.get("command/getGlobalDlLimit")
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
	resp, err := c.get("command/getGlobalUpLimit")
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
	return c.postWithHeaders("command/setPreferences", params)
}

//GetAlternativeSpeedStatus gets the alternative speed status of your qbittorrent client
func (c *Client) GetAlternativeSpeedStatus() (status bool, err error) {
	var s bool
	resp, err := c.get("command/alternativeSpeedLimitsEnabled")
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

//ToggleAlternativeSpeed toggles the alternative speed of your qbittorrent client
func (c *Client) ToggleAlternativeSpeed() (*http.Response, error) {
	return c.get("command/toggleAlternativeSpeedLimits")
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
