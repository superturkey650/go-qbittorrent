package main

import (
	"fmt"
	"go-qbittorrent/qbt"
	"go-qbittorrent/tools"
)

func main() {
	// connect to qbittorrent client
	qb := qbt.NewClient("http://localhost:8080/")

	// login to the client
	_, err := qb.Login("username", "password")
	if err != nil {
		fmt.Println(err)
	}

	// were no using an filters so the options map is empty
	options := map[string]string{}
	// set the path to the file
	file := "/Users/me/Downloads/Source.Code.2011.1080p.BluRay.H264.AAC-RARBG-[rarbg.to].torrent"
	// download the torrent using the file
	// the wrapper will handle opening and closing the file for you
	resp, err := qb.DownloadFromFile(file, options)
	if err != nil {
		tools.PrintResponse(resp.Body)
		fmt.Println(err)
	} else {
		fmt.Println("downloaded successful")
	}
}
