package qbit

//BasicTorrent holds a basic torrent object from qbittorrent
type BasicTorrent struct {
	AddedOn       int    `json:"added_on"`
	Category      string `json:"category"`
	CompletionOn  int64  `json:"completion_on"`
	Dlspeed       int    `json:"dlspeed"`
	Eta           int    `json:"eta"`
	ForceStart    bool   `json:"force_start"`
	Hash          string `json:"hash"`
	Name          string `json:"name"`
	NumComplete   int    `json:"num_complete"`
	NumIncomplete int    `json:"num_incomplete"`
	NumLeechs     int    `json:"num_leechs"`
	NumSeeds      int    `json:"num_seeds"`
	Priority      int    `json:"priority"`
	Progress      int    `json:"progress"`
	Ratio         int    `json:"ratio"`
	SavePath      string `json:"save_path"`
	SeqDl         bool   `json:"seq_dl"`
	Size          int    `json:"size"`
	State         string `json:"state"`
	SuperSeeding  bool   `json:"super_seeding"`
	Upspeed       int    `json:"upspeed"`
}

//Torrent holds a torrent object from qbittorrent
type Torrent struct {
	AdditionDate           int     `json:"addition_date"`
	Comment                string  `json:"comment"`
	CompletionDate         int     `json:"completion_date"`
	CreatedBy              string  `json:"created_by"`
	CreationDate           int     `json:"creation_date"`
	DlLimit                int     `json:"dl_limit"`
	DlSpeed                int     `json:"dl_speed"`
	DlSpeedAvg             int     `json:"dl_speed_avg"`
	Eta                    int     `json:"eta"`
	LastSeen               int     `json:"last_seen"`
	NbConnections          int     `json:"nb_connections"`
	NbConnectionsLimit     int     `json:"nb_connections_limit"`
	Peers                  int     `json:"peers"`
	PeersTotal             int     `json:"peers_total"`
	PieceSize              int     `json:"piece_size"`
	PiecesHave             int     `json:"pieces_have"`
	PiecesNum              int     `json:"pieces_num"`
	Reannounce             int     `json:"reannounce"`
	SavePath               string  `json:"save_path"`
	SeedingTime            int     `json:"seeding_time"`
	Seeds                  int     `json:"seeds"`
	SeedsTotal             int     `json:"seeds_total"`
	ShareRatio             float64 `json:"share_ratio"`
	TimeElapsed            int     `json:"time_elapsed"`
	TotalDownloaded        int     `json:"total_downloaded"`
	TotalDownloadedSession int     `json:"total_downloaded_session"`
	TotalSize              int     `json:"total_size"`
	TotalUploaded          int     `json:"total_uploaded"`
	TotalUploadedSession   int     `json:"total_uploaded_session"`
	TotalWasted            int     `json:"total_wasted"`
	UpLimit                int     `json:"up_limit"`
	UpSpeed                int     `json:"up_speed"`
	UpSpeedAvg             int     `json:"up_speed_avg"`
}

//Tracker holds a tracker object from qbittorrent
type Tracker struct {
	Msg      string `json:"msg"`
	NumPeers int    `json:"num_peers"`
	Status   string `json:"status"`
	URL      string `json:"url"`
}

//WebSeed holds a webseed object from qbittorrent
type WebSeed struct {
	URL string `json:"url"`
}

//TorrentFile holds a torrent file object from qbittorrent
type TorrentFile struct {
	IsSeed   bool   `json:"is_seed"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Progress int    `json:"progress"`
	Size     int    `json:"size"`
}

//Sync holds the sync response struct
type Sync struct {
	Categories  []string `json:"categories"`
	FullUpdate  bool     `json:"full_update"`
	Rid         int      `json:"rid"`
	ServerState struct {
		ConnectionStatus  string `json:"connection_status"`
		DhtNodes          int    `json:"dht_nodes"`
		DlInfoData        int    `json:"dl_info_data"`
		DlInfoSpeed       int    `json:"dl_info_speed"`
		DlRateLimit       int    `json:"dl_rate_limit"`
		Queueing          bool   `json:"queueing"`
		RefreshInterval   int    `json:"refresh_interval"`
		UpInfoData        int    `json:"up_info_data"`
		UpInfoSpeed       int    `json:"up_info_speed"`
		UpRateLimit       int    `json:"up_rate_limit"`
		UseAltSpeedLimits bool   `json:"use_alt_speed_limits"`
	} `json:"server_state"`
	Torrents map[string]Torrent `json:"torrents"`
}
