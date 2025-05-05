package qbt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"

	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const (
	apiBase = "api/v2/"
)

// delimit puts list into a combined (single element) map with all items connected separated by the delimiter
// this is how the WEBUI API recognizes multiple items
func delimit(items []string, delimiter string) (delimited string) {
	for i, v := range items {
		if i > 0 {
			delimited += delimiter + v
		} else {
			delimited = v
		}
	}
	return delimited
}

// Client creates a connection to qbittorrent and performs requests
type Client struct {
	http          *http.Client
	URL           string
	Authenticated bool
	Jar           http.CookieJar
}

// NewClient creates a new client connection to qbittorrent
func NewClient(url string) *Client {
	c := &Client{}

	// ensure url ends with "/"
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	c.URL = url

	// create cookie jar
	c.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c.http = &http.Client{
		Jar: c.Jar,
	}
	return c
}

// get will perform a GET request with no parameters
func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
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

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// post will perform a POST request with no content-type specified
func (c *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {

	// add optional parameters that the user wants
	form := url.Values{}
	for k, v := range opts {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		c.URL+endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.2")
	// add referer header to allow qbittorrent to identify us
	req.Header.Set("Referer", c.URL)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil

}

// postMultipart will perform a multiple part POST request
func (c *Client) postMultipart(endpoint string, buffer bytes.Buffer, contentType string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", c.URL+endpoint, &buffer)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", contentType)
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.2")

	resp, err = c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// writeOptions will write a map to the buffer through multipart.NewWriter
func writeOptions(writer *multipart.Writer, opts map[string]string) (err error) {
	for key, val := range opts {
		if err := writer.WriteField(key, val); err != nil {
			return err
		}
	}
	return nil
}

// postMultipartData will perform a multiple part POST request without a file
func (c *Client) postMultipartData(endpoint string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// write the options to the buffer
	// will contain the link string
	if err := writeOptions(writer, opts); err != nil {
		return nil, fmt.Errorf("failed to write options: %w", err)
	}

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	resp, err := c.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// postMultipartFile will perform a multiple part POST request with a file
func (c *Client) postMultipartFile(endpoint string, fileName string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	// defer the closing of the file until the end of function
	// so we can still copy its contents
	defer file.Close()

	// create form for writing the file to and give it the filename
	formWriter, err := writer.CreateFormFile("torrents", path.Base(fileName))
	if err != nil {
		return nil, fmt.Errorf("error adding file: %w", err)
	}

	// write the options to the buffer
	writeOptions(writer, opts)

	// copy the file contents into the form
	if _, err = io.Copy(formWriter, file); err != nil {
		return nil, fmt.Errorf("error copying file: %w", err)
	}

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	resp, err := c.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Application endpoints

// Login authenticates with the qBittorrent client using the provided credentials.
// It returns an error if authentication fails or if the client's IP is banned.
func (c *Client) Login(opts LoginOptions) (err error) {
	params := map[string]string{}

	if opts.Username != "" {
		params["username"] = opts.Username
	}
	if opts.Password != "" {
		params["password"] = opts.Password
	}

	if c.http == nil {
		c.http = &http.Client{Jar: c.Jar}
	}

	resp, err := c.post(apiBase+"auth/login", params)
	if err != nil {
		return err
	} else if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("user's IP is banned for too many failed login attempts: %w", err)
	}

	// change authentication status so we know were authenticated in later requests
	c.Authenticated = true

	return nil
}

// Logout logs you out of the qbittorrent client
// returns the current authentication status
func (c *Client) Logout() (err error) {
	resp, err := c.get(apiBase+"auth/logout", nil)
	if err != nil {
		return err
	}

	// change authentication status so we know were not authenticated in later requests
	c.Authenticated = (*resp).StatusCode == 200

	return nil
}

// ApplicationVersion of the qbittorrent client
func (c *Client) ApplicationVersion() (version string, err error) {
	resp, err := c.get(apiBase+"app/version", nil)
	if err != nil {
		return version, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return version, err
	}
	return version, err
}

// WebAPIVersion of the qbittorrent client
func (c *Client) WebAPIVersion() (version string, err error) {
	resp, err := c.get(apiBase+"app/webapiVersion", nil)
	if err != nil {
		return version, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return version, err
	}
	return version, err
}

// BuildInfo of the qbittorrent client
func (c *Client) BuildInfo() (buildInfo BuildInfo, err error) {
	resp, err := c.get(apiBase+"app/buildInfo", nil)
	if err != nil {
		return buildInfo, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&buildInfo); err != nil {
		return buildInfo, err
	}
	return buildInfo, err
}

// Preferences of the qbittorrent client
func (c *Client) Preferences() (prefs Preferences, err error) {
	resp, err := c.get(apiBase+"app/preferences", nil)
	if err != nil {
		return prefs, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return prefs, err
	}
	return prefs, err
}

// SetPreferences of the qbittorrent client
func (c *Client) SetPreferences() (prefsSet bool, err error) {
	resp, err := c.post(apiBase+"app/setPreferences", nil)
	return (resp.StatusCode == http.StatusOK), err
}

// DefaultSavePath of the qbittorrent client
func (c *Client) DefaultSavePath() (path string, err error) {
	resp, err := c.get(apiBase+"app/defaultSavePath", nil)
	if err != nil {
		return path, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&path); err != nil {
		return path, err
	}
	return path, err
}

// Shutdown shuts down the qbittorrent client
func (c *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := c.post(apiBase+"app/shutdown", nil)
	if err != nil {
		return shuttingDown, err
	}
	// return true if successful
	return (resp.StatusCode == http.StatusOK), err
}

// Log Endpoints

// Logs of the qbittorrent client
func (c *Client) Logs(filters map[string]string) (logs []Log, err error) {
	resp, err := c.get(apiBase+"log/main", filters)
	if err != nil {
		return logs, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return logs, err
	}
	return logs, err
}

// PeerLogs of the qbittorrent client
func (c *Client) PeerLogs(filters map[string]string) (logs []PeerLog, err error) {
	resp, err := c.get(apiBase+"log/peers", filters)
	if err != nil {
		return logs, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return logs, err
	}
	return logs, err
}

// Sync Endpoints

// MainData returns info you usually see in qBt status bar.
func (c *Client) MainData(opts MainDataOptions) (mainData MainData, err error) {
	resp, err := c.get(apiBase+"sync/mainData", nil)
	if err != nil {
		return mainData, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&mainData); err != nil {
		return mainData, err
	}
	return mainData, err
}

// TorrentPeers returns info you usually see in qBt status bar.
func (c *Client) TorrentPeers(opts TorrentPeersOptions) (torrentPeers TorrentPeers, err error) {
	resp, err := c.get(apiBase+"sync/mainData", nil)
	if err != nil {
		return torrentPeers, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&torrentPeers); err != nil {
		return torrentPeers, err
	} else if resp != nil && (*resp).StatusCode == http.StatusNotFound {
		return torrentPeers, fmt.Errorf("torrent hash not found")
	}
	return torrentPeers, err
}

// TODO: Transfer Endpoints

// Info returns info you usually see in qBt status bar.
func (c *Client) Info(opts InfoOptions) (info Info, err error) {
	resp, err := c.get(apiBase+"transfer/info", nil)
	if err != nil {
		return info, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return info, err
	}
	return info, err
}

// AltSpeedLimitsEnabled returns info you usually see in qBt status bar.
func (c *Client) AltSpeedLimitsEnabled() (mode bool, err error) {
	resp, err := c.get(apiBase+"transfer/speedLimitsMode", nil)
	if err != nil {
		return mode, err
	}
	var decoded int
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return mode, err
	}
	mode = decoded == 1
	return mode, err
}

// ToggleAltSpeedLimits returns info you usually see in qBt status bar.
func (c *Client) ToggleAltSpeedLimits() (toggled bool, err error) {
	resp, err := c.get(apiBase+"transfer/toggleSpeedLimitsMode", nil)
	if err != nil {
		return toggled, err
	}
	return (resp.StatusCode == http.StatusOK), err
}

// DlLimit returns info you usually see in qBt status bar.
func (c *Client) DlLimit() (dlLimit int, err error) {
	resp, err := c.get(apiBase+"transfer/downloadLimit", nil)
	if err != nil {
		return dlLimit, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&dlLimit); err != nil {
		return dlLimit, err
	}
	return dlLimit, err
}

// SetDlLimit returns info you usually see in qBt status bar.
func (c *Client) SetDlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := c.get(apiBase+"transfer/setDownloadLimit", params)
	if err != nil {
		return set, err
	}
	return (resp.StatusCode == http.StatusOK), err
}

// UlLimit returns info you usually see in qBt status bar.
func (c *Client) UlLimit() (ulLimit int, err error) {
	resp, err := c.get(apiBase+"transfer/uploadLimit", nil)
	if err != nil {
		return ulLimit, err
	}
	json.NewDecoder(resp.Body).Decode(&ulLimit)
	return ulLimit, err
}

// SetUlLimit returns info you usually see in qBt status bar.
func (c *Client) SetUlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := c.get(apiBase+"transfer/setUploadLimit", params)
	if err != nil {
		return set, err
	}
	return (resp.StatusCode == http.StatusOK), err
}

// Torrents returns a list of all torrents in qbittorrent matching your filter
func (c *Client) Torrents(opts TorrentsOptions) (torrentList []TorrentInfo, err error) {
	params := map[string]string{}
	if opts.Filter != nil {
		params["filter"] = *opts.Filter
	}
	if opts.Category != nil {
		params["category"] = *opts.Category
	}
	if opts.Sort != nil {
		params["sort"] = *opts.Sort
	}
	if opts.Reverse != nil {
		params["reverse"] = strconv.FormatBool(*opts.Reverse)
	}
	if opts.Offset != nil {
		params["offset"] = strconv.Itoa(*opts.Offset)
	}
	if opts.Limit != nil {
		params["limit"] = strconv.Itoa(*opts.Limit)
	}
	if opts.Hashes != nil {
		params["hashes"] = delimit(opts.Hashes, "%0A")
	}
	resp, err := c.get(apiBase+"torrents/info", params)
	if err != nil {
		return torrentList, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&torrentList); err != nil {
		return torrentList, err
	}
	return torrentList, nil
}

// Torrent returns a specific torrent matching the hash
func (c *Client) Torrent(hash string) (torrent Torrent, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/properties", opts)
	if err != nil {
		return torrent, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&torrent); err != nil {
		return torrent, err
	}
	return torrent, nil
}

// TorrentTrackers returns all trackers for a specific torrent matching the hash
func (c *Client) TorrentTrackers(hash string) (trackers []Tracker, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/trackers", opts)
	if err != nil {
		return trackers, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&trackers); err != nil {
		return trackers, err
	}
	return trackers, nil
}

// TorrentWebSeeds returns seeders for a specific torrent matching the hash
func (c *Client) TorrentWebSeeds(hash string) (webSeeds []WebSeed, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/webseeds", opts)
	if err != nil {
		return webSeeds, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&webSeeds); err != nil {
		return webSeeds, err
	}
	return webSeeds, nil
}

// TorrentFiles from given hash
func (c *Client) TorrentFiles(hash string) (files []TorrentFile, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/files", opts)
	if err != nil {
		return files, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return files, err
	}
	return files, nil
}

// TorrentPieceStates for all pieces of torrent
func (c *Client) TorrentPieceStates(hash string) (states []int, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/pieceStates", opts)
	if err != nil {
		return states, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return states, err
	}
	return states, nil
}

// TorrentPieceHashes for all pieces of torrent
func (c *Client) TorrentPieceHashes(hash string) (hashes []string, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := c.get(apiBase+"torrents/pieceHashes", opts)
	if err != nil {
		return hashes, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&hashes); err != nil {
		return hashes, err
	}
	return hashes, nil
}

// Pause torrents
func (c *Client) Pause(hashes []string) error {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	_, err := c.get(apiBase+"torrents/pause", opts)
	if err != nil {
		return err
	}

	return nil
}

// Resume torrents
func (c *Client) Resume(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/resume", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// Delete torrents and optionally delete their files
func (c *Client) Delete(hashes []string, deleteFiles bool) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	opts["deleteFiles"] = strconv.FormatBool(deleteFiles)
	resp, err := c.get(apiBase+"torrents/delete", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// Recheck torrents
func (c *Client) Recheck(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/recheck", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// Reannounce torrents
func (c *Client) Reannounce(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/reannounce", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// DownloadFromLink starts downloading a torrent from a link
func (c *Client) DownloadLinks(links []string, opts DownloadOptions) error {
	params := map[string]string{}
	if len(links) == 0 {
		return fmt.Errorf("at least one url must be present")
	} else {
		delimitedURLs := delimit(links, "%0A")
		// TODO: Why is encoding causing problems now?
		// encodedURLS := url.QueryEscape(delimitedURLs)
		params["urls"] = delimitedURLs
	}
	if opts.Savepath != nil {
		params["savepath"] = *opts.Savepath
	}
	if opts.Cookie != nil {
		params["cookie"] = *opts.Cookie
	}
	if opts.Category != nil {
		params["category"] = *opts.Category
	}
	if opts.SkipHashChecking != nil {
		params["skip_checking"] = strconv.FormatBool(*opts.SkipHashChecking)
	}
	if opts.Paused != nil {
		params["paused"] = strconv.FormatBool(*opts.Paused)
	}
	if opts.RootFolder != nil {
		params["root_folder"] = strconv.FormatBool(*opts.RootFolder)
	}
	if opts.Rename != nil {
		params["rename"] = *opts.Rename
	}
	if opts.UploadSpeedLimit != nil {
		params["upLimit"] = strconv.Itoa(*opts.UploadSpeedLimit)
	}
	if opts.DownloadSpeedLimit != nil {
		params["dlLimit"] = strconv.Itoa(*opts.DownloadSpeedLimit)
	}
	if opts.SequentialDownload != nil {
		params["sequentialDownload"] = strconv.FormatBool(*opts.SequentialDownload)
	}
	if opts.FirstLastPiecePriority != nil {
		params["firstLastPiecePrio"] = strconv.FormatBool(*opts.FirstLastPiecePriority)
	}

	resp, err := c.postMultipartData(apiBase+"torrents/add", params)
	if err != nil {
		return err
	} else if resp.StatusCode == 415 {
		return fmt.Errorf("torrent file is not valid")
	}

	return nil
}

// DownloadFromFile starts downloading a torrent from a file
func (c *Client) DownloadFromFile(torrents string, opts DownloadOptions) error {
	params := map[string]string{}
	if torrents == "" {
		return fmt.Errorf("at least one file must be present")
	}
	if opts.Savepath != nil {
		params["savepath"] = *opts.Savepath
	}
	if opts.Cookie != nil {
		params["cookie"] = *opts.Cookie
	}
	if opts.Category != nil {
		params["category"] = *opts.Category
	}
	if opts.SkipHashChecking != nil {
		params["skip_checking"] = strconv.FormatBool(*opts.SkipHashChecking)
	}
	if opts.Paused != nil {
		params["paused"] = strconv.FormatBool(*opts.Paused)
	}
	if opts.RootFolder != nil {
		params["root_folder"] = strconv.FormatBool(*opts.RootFolder)
	}
	if opts.Rename != nil {
		params["rename"] = *opts.Rename
	}
	if opts.UploadSpeedLimit != nil {
		params["upLimit"] = strconv.Itoa(*opts.UploadSpeedLimit)
	}
	if opts.DownloadSpeedLimit != nil {
		params["dlLimit"] = strconv.Itoa(*opts.DownloadSpeedLimit)
	}
	if opts.AutomaticTorrentManagement != nil {
		params["autoTMM"] = strconv.FormatBool(*opts.AutomaticTorrentManagement)
	}
	if opts.SequentialDownload != nil {
		params["sequentialDownload"] = strconv.FormatBool(*opts.SequentialDownload)
	}
	if opts.FirstLastPiecePriority != nil {
		params["firstLastPiecePrio"] = strconv.FormatBool(*opts.FirstLastPiecePriority)
	}
	resp, err := c.postMultipartFile(apiBase+"torrents/add", torrents, params)
	if err != nil {
		return err
	} else if resp.StatusCode == 415 {
		return fmt.Errorf("torrent file is not valid")
	}

	return nil
}

// AddTrackers to a torrent
func (c *Client) AddTrackers(hash string, trackers []string) error {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(hash)
	delimitedTrackers := delimit(trackers, "%0A")
	encodedTrackers := url.QueryEscape(delimitedTrackers)
	params["urls"] = encodedTrackers

	resp, err := c.post(apiBase+"torrents/addTrackers", params)
	if err != nil {
		return err
	} else if resp != nil && (*resp).StatusCode == http.StatusNotFound {
		return fmt.Errorf("torrent hash not found")
	}
	return nil
}

// EditTracker on a torrent
func (c *Client) EditTracker(hash string, origURL string, newURL string) error {
	params := map[string]string{
		"hash":    hash,
		"origUrl": origURL,
		"newUrl":  newURL,
	}
	resp, err := c.get(apiBase+"torrents/editTracker", params)
	if err != nil {
		return err
	}
	switch sc := (*resp).StatusCode; sc {
	case http.StatusBadRequest:
		return fmt.Errorf("newUrl is not a valid url")
	case http.StatusNotFound:
		return fmt.Errorf("torrent hash was not found")
	case http.StatusConflict:
		return fmt.Errorf("newUrl already exists for this torrent or origUrl was not found")
	default:
		return nil
	}
}

// RemoveTrackers from a torrent
func (c *Client) RemoveTrackers(hash string, trackers []string) error {
	params := map[string]string{
		"hash": hash,
		"urls": delimit(trackers, "|"),
	}
	resp, err := c.get(apiBase+"torrents/removeTrackers", params)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("torrent hash was not found")
	case http.StatusConflict:
		return fmt.Errorf("all URLs were not found")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// IncreasePriority of torrents
func (c *Client) IncreasePriority(hashes []string) error {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/IncreasePrio", opts)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("torrent queueing is not enabled")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// DecreasePriority of torrents
func (c *Client) DecreasePriority(hashes []string) error {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/DecreasePrio", opts)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("Torrent queueing is not enabled")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// MaxPriority maximizes the priority of torrents
func (c *Client) MaxPriority(hashes []string) error {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/TopPrio", opts)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("torrent queueing is not enabled")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// MinPriority maximizes the priority of torrents
func (c *Client) MinPriority(hashes []string) error {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/BottomPrio", opts)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("torrent queueing is not enabled")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// FilePriority for a torrent
func (c *Client) FilePriority(hash string, ids []int, priority int) error {
	formattedIds := []string{}
	for _, id := range ids {
		formattedIds = append(formattedIds, strconv.Itoa(id))
	}

	opts := map[string]string{
		"hashes":   hash,
		"id":       delimit(formattedIds, "|"),
		"priority": strconv.Itoa(priority),
	}
	resp, err := c.get(apiBase+"torrents/filePrio", opts)
	if err != nil {
		return err
	}

	switch sc := (*resp).StatusCode; sc {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("priority is invalid or at least one id is not an integer")
	case http.StatusConflict:
		return fmt.Errorf("Torrent metadata hasn't downloaded yet or at least one file id was not found")
	default:
		return fmt.Errorf("an unknown error occurred causing a status code of: %d", sc)
	}
}

// GetTorrentDownloadLimit for a list of torrents
func (c *Client) GetTorrentDownloadLimit(hashes []string) (limits map[string]int, err error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.post(apiBase+"torrents/downloadLimit", opts)
	if err != nil {
		return limits, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return limits, err
	}
	return limits, nil
}

// SetTorrentDownloadLimit for a list of torrents
func (c *Client) SetTorrentDownloadLimit(hashes []string, limit int) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"limit":  strconv.Itoa(limit),
	}
	resp, err := c.post(apiBase+"torrents/setDownloadLimit", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// SetTorrentShareLimit for a list of torrents
func (c *Client) SetTorrentShareLimit(hashes []string, ratioLimit int, seedingTimeLimit int) (bool, error) {
	opts := map[string]string{
		"hashes":           delimit(hashes, "|"),
		"ratioLimit":       strconv.Itoa(ratioLimit),
		"seedingTimeLimit": strconv.Itoa(seedingTimeLimit),
	}
	resp, err := c.post(apiBase+"torrents/setShareLimits", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// GetTorrentUploadLimit for a list of torrents
func (c *Client) GetTorrentUploadLimit(hashes []string) (limits map[string]int, err error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.post(apiBase+"torrents/uploadLimit", opts)
	if err != nil {
		return limits, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return limits, err
	}
	return limits, nil
}

// SetTorrentUploadLimit for a list of torrents
func (c *Client) SetTorrentUploadLimit(hashes []string, limit int) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"limit":  strconv.Itoa(limit),
	}
	resp, err := c.post(apiBase+"torrents/setUploadLimit", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

// SetTorrentLocation for a list of torrents
func (c *Client) SetTorrentLocation(hashes []string, location string) (bool, error) {
	opts := map[string]string{
		"hashes":   delimit(hashes, "|"),
		"location": location,
	}
	resp, err := c.post(apiBase+"torrents/setLocation", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// SetTorrentName for a torrent
func (c *Client) SetTorrentName(hash string, name string) (bool, error) {
	opts := map[string]string{
		"hash": hash,
		"name": name,
	}
	resp, err := c.post(apiBase+"torrents/rename", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// SetTorrentCategory for a list of torrents
func (c *Client) SetTorrentCategory(hashes []string, category string) (bool, error) {
	opts := map[string]string{
		"hashes":   delimit(hashes, "|"),
		"category": category,
	}
	resp, err := c.post(apiBase+"torrents/setCategory", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// GetCategories used by client
func (c *Client) GetCategories() (categories Categories, err error) {
	resp, err := c.get(apiBase+"torrents/categories", nil)
	if err != nil {
		return categories, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return categories, err
	}
	return categories, nil
}

// CreateCategory for use by client
func (c *Client) CreateCategory(category string, savePath string) (bool, error) {
	opts := map[string]string{
		"category": category,
		"savePath": savePath,
	}
	resp, err := c.post(apiBase+"torrents/createCategory", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// UpdateCategory used by client
func (c *Client) UpdateCategory(category string, savePath string) (bool, error) {
	opts := map[string]string{
		"category": category,
		"savePath": savePath,
	}
	resp, err := c.post(apiBase+"torrents/editCategory", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// DeleteCategories used by client
func (c *Client) DeleteCategories(categories []string) (bool, error) {
	opts := map[string]string{"categories": delimit(categories, "\n")}
	resp, err := c.post(apiBase+"torrents/removeCategories", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// AddTorrentTags to a list of torrents
func (c *Client) AddTorrentTags(hashes []string, tags []string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"tags":   delimit(tags, ","),
	}
	resp, err := c.post(apiBase+"torrents/addTags", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// RemoveTorrentTags from a list of torrents (empty list removes all tags)
func (c *Client) RemoveTorrentTags(hashes []string, tags []string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"tags":   delimit(tags, ","),
	}
	resp, err := c.post(apiBase+"torrents/removeTags", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// GetTorrentTags from a list of torrents (empty list removes all tags)
func (c *Client) GetTorrentTags() (tags []string, err error) {
	resp, err := c.get(apiBase+"torrents/tags", nil)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return tags, err
	}
	return tags, nil
}

// CreateTags for use by client
func (c *Client) CreateTags(tags []string) (bool, error) {
	opts := map[string]string{"tags": delimit(tags, ",")}
	resp, err := c.post(apiBase+"torrents/createTags", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// DeleteTags used by client
func (c *Client) DeleteTags(tags []string) (bool, error) {
	opts := map[string]string{"tags": delimit(tags, ",")}
	resp, err := c.post(apiBase+"torrents/deleteTags", opts)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// SetAutoManagement for a list of torrents
func (c *Client) SetAutoManagement(hashes []string, enable bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"enable": strconv.FormatBool(enable),
	}
	resp, err := c.post(apiBase+"torrents/setAutoManagement", opts)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// ToggleSequentialDownload for a list of torrents
func (c *Client) ToggleSequentialDownload(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/toggleSequentialDownload", opts)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// ToggleFirstLastPiecePriority for a list of torrents
func (c *Client) ToggleFirstLastPiecePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := c.get(apiBase+"torrents/toggleFirstLastPiecePrio", opts)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// SetForceStart for a list of torrents
func (c *Client) SetForceStart(hashes []string, value bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"value":  strconv.FormatBool(value),
	}
	resp, err := c.post(apiBase+"torrents/setForceStart", opts)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}

// SetSuperSeeding for a list of torrents
func (c *Client) SetSuperSeeding(hashes []string, value bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"value":  strconv.FormatBool(value),
	}
	resp, err := c.post(apiBase+"torrents/setSuperSeeding", opts)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil //TODO: look into other statuses
}
