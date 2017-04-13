package tools

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

//PrintResponse prints the body of a response
func PrintResponse(body io.ReadCloser) {
	r := make([]byte, 256)
	r, _ = ioutil.ReadAll(body)
	fmt.Println("response: " + string(r))
}

//PrintRequest prints a request
func PrintRequest(req *http.Request) {
	r, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("request: " + string(r))
}
