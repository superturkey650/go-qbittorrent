package qbt

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

func PrintResponse(body io.ReadCloser) {
	r := make([]byte, 256)
	r, _ = ioutil.ReadAll(body)
	fmt.Println("response: " + string(r))
}

func PrintRequest(req *http.Request) {
	r, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("request: " + string(r))
}
