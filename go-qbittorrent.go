package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func main() {
	// connect to qbittorrent client
	qb := qbt.NewClient("http://localhost:8080")

	// login to the client
	loginOpts := qbt.LoginOptions{
		Username: "username",
		Password: "password",
	}
	err := qb.Login(loginOpts)
	if err != nil {
		fmt.Println(err)
	}

	// ********************
	// DOWNLOAD A TORRENT *
	// ********************

	// were not using any filters so the options map is empty
	downloadOpts := qbt.DownloadOptions{}
	// set the path to the file
	//path := "/Users/me/Downloads/Source.Code.2011.1080p.BluRay.H264.AAC-RARBG-[rarbg.to].torrent"
	links := []string{"magnet:?xt=urn:btih:7CDD7029DA5AEAF2589375C32BFC0A065A8AE3A7&dn=Courage%20The%20Cowardly%20Dog%20S01E08%20480p%20x264-mSD&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=udp%3A%2F%2Ftracker.bittor.pw%3A1337%2Fannounce&tr=udp%3A%2F%2Fpublic.popcorn-tracker.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.dler.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce"}
	// download the torrent using the file
	// the wrapper will handle opening and closing the file for you
	err = qb.DownloadLinks(links, downloadOpts)

	if err != nil {
		fmt.Println("[-] Download torrent from link")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Download torrent from link")
	}

	// ******************
	// GET ALL TORRENTS *
	// ******************

	torrentsOpts := qbt.TorrentsOptions{}
	//filter := "inactive"
	sort := "name"
	//hash := "d739f78a12b241ba62719b1064701ffbb45498a8"
	//torrentsOpts.Filter = &filter
	torrentsOpts.Sort = &sort
	//torrentsOpts.Hashes = []string{hash}
	torrents, err := qb.Torrents(torrentsOpts)
	if err != nil {
		fmt.Println("[-] Get torrent list")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Get torrent list")
		if len(torrents) > 0 {
			spew.Dump(torrents[0])
		} else {
			fmt.Println("No torrents found")
		}
	}
}
