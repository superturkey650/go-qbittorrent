package main

import (
	"fmt"
	"go-qbittorrent(extra)/qbit"
	"time"
)

func main() {
	qb := qbit.NewClient("http://localhost:8080/")
	//time.Sleep(time.Second)
	qb.Login("pavo", "buffalo12")
	time.Sleep(time.Second * 2)
	options := make(map[string]string)
	file := "/Users/me/Downloads/Source.Code.2011.1080p.BluRay.H264.AAC-RARBG-[rarbg.to].torrent"
	_, err := qb.DownloadFromFile(file, options)
	if err != nil {
		fmt.Println("error on downloading from file")
		fmt.Println(err)
	} else {
		fmt.Println("success")
	}
}
