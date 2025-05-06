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

	// set up a test hash to pause and resume
	testHashes := []string{"7cdd7029da5aeaf2589375c32bfc0a065a8ae3a7"}

	// *****************
	// PAUSE A TORRENT *
	// *****************
	err = qb.Pause(testHashes)
	if err != nil {
		fmt.Println("[-] Pause torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Pause torrent")
	}

	// ******************
	// RESUME A TORRENT *
	// ******************
	err = qb.Resume(testHashes)
	if err != nil {
		fmt.Println("[-] Resume torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Resume torrent")
	}

	// ******************
	// DELETE A TORRENT *
	// ******************
	// err = qb.Delete(testHashes, true)
	// if err != nil {
	// 	fmt.Println("[-] Delete torrent")
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("[+] Delete torrent")
	// }

	appVersion, err := qb.ApplicationVersion()
	if err != nil {
		fmt.Println("[-] Application Version")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Application Version: ", appVersion)
	}

	webAPIVersion, err := qb.WebAPIVersion()
	if err != nil {
		fmt.Println("[-] Web API Version")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Web API Version: ", webAPIVersion)
	}

	buildInfo, err := qb.BuildInfo()
	if err != nil {
		fmt.Println("[-] Build Info")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Build Info: ", buildInfo)
	}

	preferences, err := qb.Preferences()
	if err != nil {
		fmt.Println("[-] Preferences")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Preferences: ", preferences)
	}

	// TODO: Preferences

	defaultSavePath, err := qb.DefaultSavePath()
	if err != nil {
		fmt.Println("[-] Default Save Path")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Default Save Path: ", defaultSavePath)
	}

	// LOG ENDPOINTS

	logs, err := qb.Logs(nil)
	if err != nil {
		fmt.Println("[-] Logs")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Logs: ", logs[34])
	}

	peerLogs, err := qb.PeerLogs(nil)
	if err != nil {
		fmt.Println("[-] Peer Logs")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Peer Logs: ", peerLogs)
	}

	// SYNC ENDPOINTS

	mainData, err := qb.MainData("")
	if err != nil {
		fmt.Println("[-] Main Data")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Main Data: ", mainData)
	}

	torrentPeers, err := qb.TorrentPeers(testHashes[0], "")
	if err != nil {
		fmt.Println("[-] Torrent Peers")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Peers: ", torrentPeers)
	}

	// TRANSFER ENDPOINTS

	info, err := qb.Info()
	if err != nil {
		fmt.Println("[-] Info")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Info: ", info)
	}

	altSpeedLimitsEnabled, err := qb.AltSpeedLimitsEnabled()
	if err != nil {
		fmt.Println("[-] Alt Speed Limits Enabled")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Alt Speed Limits Enabled: ", altSpeedLimitsEnabled)
	}

	// err = qb.ToggleAltSpeedLimits()
	// if err != nil {
	// 	fmt.Println("[-] Toggled Alt Speed Limits")
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("[+] Toggled Alt Speed Limits")
	// }

	dlLimit, err := qb.DlLimit()
	if err != nil {
		fmt.Println("[-] Download Limit")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Download Limit: ", dlLimit)
	}

	err = qb.SetDlLimit(0)
	if err != nil {
		fmt.Println("[-] Set Download Limit")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Set Download Limit")
	}

	ulLimit, err := qb.UlLimit()
	if err != nil {
		fmt.Println("[-] Upload Limit")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Upload Limit: ", ulLimit)
	}

	err = qb.SetUlLimit(0)
	if err != nil {
		fmt.Println("[-] Set Upload Limit")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Set Upload Limit")
	}

	// TODO: Ban Peers
}
