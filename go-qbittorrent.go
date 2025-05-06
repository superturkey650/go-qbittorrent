package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/superturkey650/go-qbittorrent/qbt"
)

func ptrString(s string) *string {
	return &s
}

func main() {
	// set up a test hash to pause and resume
	testHashes := []string{"7cdd7029da5aeaf2589375c32bfc0a065a8ae3a7"}

	// connect to qbittorrent client
	qb := qbt.NewClient("http://localhost:8080")

	// Application Endpoints

	// login to the client
	loginOpts := qbt.LoginOptions{
		Username: "username",
		Password: "password",
	}
	err := qb.Login(loginOpts)
	if err != nil {
		fmt.Println(err)
	}

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

	// Torrent Endpoints

	// ******************
	// GET ALL TORRENTS *
	// ******************

	//filter := "inactive"
	//hash := "d739f78a12b241ba62719b1064701ffbb45498a8"
	//torrentsOpts.Filter = &filter
	//torrentsOpts.Hashes = []string{hash}
	torrentsOpts := qbt.TorrentsOptions{
		Sort: ptrString("name"),
	}
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

	torrent, err := qb.Torrent(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent: ", torrent)
	}

	torrentTrackers, err := qb.TorrentTrackers(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent Trackers")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Trackers: ", torrentTrackers)
	}

	torrentWebSeeds, err := qb.TorrentWebSeeds(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent Web Seeds")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Web Seeds: ", torrentWebSeeds)
	}

	torrentFiles, err := qb.TorrentFiles(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent Files")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Files: ", torrentFiles)
	}

	torrentPieceStates, err := qb.TorrentPieceStates(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent Piece States")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Piece States: ", torrentPieceStates)
	}

	torrentPieceHashes, err := qb.TorrentPieceHashes(testHashes[0])
	if err != nil {
		fmt.Println("[-] Torrent Piece Hashes")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Piece Hashes: ", torrentPieceHashes)
	}

	err = qb.Pause(testHashes)
	if err != nil {
		fmt.Println("[-] Pause torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Pause torrent")
	}

	err = qb.Resume(testHashes)
	if err != nil {
		fmt.Println("[-] Resume torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Resume torrent")
	}

	// err = qb.Delete(testHashes, true)
	// if err != nil {
	// 	fmt.Println("[-] Delete torrent")
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("[+] Delete torrent")
	// }

	err = qb.Recheck(testHashes)
	if err != nil {
		fmt.Println("[-] Recheck torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Recheck torrent")
	}

	err = qb.Reannounce(testHashes)
	if err != nil {
		fmt.Println("[-] Reannounce torrent")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Reannounce torrent")
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

	// TODO: File-based download

	// TODO: Tracker actions

	err = qb.IncreasePriority(testHashes)
	if err != nil {
		fmt.Println("[-] Increase Priority")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Increase Priority")
	}

	err = qb.DecreasePriority(testHashes)
	if err != nil {
		fmt.Println("[-] Decrease Priority")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Decrease Priority")
	}

	err = qb.MaxPriority(testHashes)
	if err != nil {
		fmt.Println("[-] Max Priority")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Max Priority")
	}

	err = qb.MinPriority(testHashes)
	if err != nil {
		fmt.Println("[-] Min Priority")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Min Priority")
	}

	// TODO: File Priority

	dlLimits, err := qb.GetTorrentDownloadLimit(testHashes)
	if err != nil {
		fmt.Println("[-] Download Limits")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Download Limits: ", dlLimits)
	}

	err = qb.SetTorrentDownloadLimit(testHashes, dlLimits[testHashes[0]])
	if err != nil {
		fmt.Println("[-] Download Limits")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Download Limits")
	}

	err = qb.SetTorrentShareLimit(testHashes, 1, -2, -2)
	if err != nil {
		fmt.Println("[-] Torrent Share Limits")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Share Limits")
	}

	ulLimits, err := qb.GetTorrentUploadLimit(testHashes)
	if err != nil {
		fmt.Println("[-] Upload Limits")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Upload Limits: ", ulLimits)
	}

	err = qb.SetTorrentUploadLimit(testHashes, ulLimits[testHashes[0]])
	if err != nil {
		fmt.Println("[-] Upload Limits")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Upload Limits")
	}

	err = qb.SetTorrentLocation(testHashes, "/Users/jared/Downloads")
	if err != nil {
		fmt.Println("[-] Torrent Location")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Location")
	}

	err = qb.SetTorrentName(testHashes[0], "test_name")
	if err != nil {
		fmt.Println("[-] Torrent Name")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Name")
	}

	err = qb.SetTorrentCategory(testHashes, "")
	if err != nil {
		fmt.Println("[-] Torrent Name")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Torrent Name")
	}

	categories, err := qb.GetCategories()
	if err != nil {
		fmt.Println("[-] Categories")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Categories: ", categories)
	}

	err = qb.CreateCategory("test_category", "/Users/jared/Downloads")
	if err != nil {
		fmt.Println("[-] Create Category")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Create Category")
	}

	err = qb.UpdateCategory("test_categoria", "/Users/jared/Downloads/test")
	if err != nil {
		fmt.Println("[-] Update Category")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Update Category")
	}

	categories, err = qb.GetCategories()
	if err != nil {
		fmt.Println("[-] Categories")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Categories: ", categories)
	}

	testCategories := []string{"test_category"}
	err = qb.DeleteCategories(testCategories)
	if err != nil {
		fmt.Println("[-] Delete Category")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Delete Category")
	}

	categories, err = qb.GetCategories()
	if err != nil {
		fmt.Println("[-] Categories")
		fmt.Println(err)
	} else {
		fmt.Println("[+] Categories: ", categories)
	}

	// TODO: Tags
}
