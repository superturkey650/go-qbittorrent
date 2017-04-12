package qbit

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"path"

	"github.com/pkg/errors"
)

func (c *Client) get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL+endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

func (c *Client) getWithParams(endpoint string, params map[string]string) (*http.Response, error) {

	req, err := http.NewRequest("GET", c.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	//add parameters to url
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	req.Close = true

	return resp, nil
}

func addForm(req *http.Request, params map[string]string) *http.Request {
	form := url.Values{}
	for k, v := range params {
		form.Add(k, v)
	}
	req.PostForm = form
	return req
}

func (c *Client) post(endpoint string, data map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.URL+endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	req = addForm(req, data)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	return resp, nil

}

func (c *Client) postWithHeaders(endpoint string, data map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.URL+endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	//add headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	req = addForm(req, data)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

func (c *Client) postMultipart(endpoint string, data map[string]string) (*http.Response, error) {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, val := range data {
		w.WriteField(key, val)
	}
	contentType := w.FormDataContentType()

	err := w.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to close writer")
	}

	resp, err := http.Post(c.URL+endpoint, contentType, &b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

func (c *Client) postMultipartFile(endpoint string, data map[string]string, file string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "error opening file")
	}

	defer f.Close()

	fw, err := w.CreateFormFile("torrents", path.Base(file))
	if err != nil {
		return nil, errors.Wrap(err, "error adding file")
	}

	if _, err = io.Copy(fw, f); err != nil {
		return nil, errors.Wrap(err, "error copying file")
	}

	w.Close()

	req, err := http.NewRequest("POST", c.URL+endpoint, &b)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform request")
	}

	return resp, nil
}
