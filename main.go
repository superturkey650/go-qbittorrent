package main

import (
	"fmt"
	"go-qbittorrent/qbt"
)

func main() {
	qb := qbt.NewClient("http://localhost:8080/")

	qb.Login("username", "password")

	options := map[string]string{}
	file := "/Users/me/Downloads/Source.Code.2011.1080p.BluRay.H264.AAC-RARBG-[rarbg.to].torrent"
	_, err := qb.DownloadFromFile(file, options)
	if err != nil {
		fmt.Println("error on downloading from file")
		fmt.Println(err)
	} else {
		fmt.Println("success")
	}
}
