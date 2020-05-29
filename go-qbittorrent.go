package main

import (
	"fmt"

	"github.com/superturkey650/go-qbittorrent/qbt"
)

func main() {
	// connect to qbittorrent client
	qb := qbt.NewClient("http://localhost:8181")

	// login to the client
	loginOpts := qbt.LoginOptions{
		Username: "username",
		Password: "password",
	}
	err := qb.Login(loginOpts)
	if err != nil {
		fmt.Println(err)
	}

	// were not using any filters so the options map is empty
	options := map[string]string{}
	// set the path to the file
	//path := "/Users/me/Downloads/Source.Code.2011.1080p.BluRay.H264.AAC-RARBG-[rarbg.to].torrent"
	link := "http://rarbg.to/download.php?id=ita4eys&h=e6c&f=Avengers.Endgame.2019.2160p.BluRay.x265.10bit.SDR.DTS-HD.MA.TrueHD.7.1.Atmos-SWTYBLZ-[rarbg.to].torrent"
	// download the torrent using the file
	// the wrapper will handle opening and closing the file for you
	resp, err := qb.DownloadFromLink(link, options)

	if err != nil {
		fmt.Println(err)
	} else if resp != nil && (*resp).StatusCode != 200 {
		fmt.Println("Unable to login")
	} else {
		fmt.Println("downloaded successful")
	}
}
