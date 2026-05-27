package models

type MediaContainer struct {
	Accounts []Account `xml:"Account"`
}

type Account struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type User struct {
	Username string `xml:"username,attr"`
	Email    string `xml:"email,attr"`
	ID       string `xml:"id,attr"`
}

type DownloadData struct {
	Fid             string `json:"fid"`
	Filename        string `json:"filename"`
	CategoryID      int    `json:"categoryID"`
	Size            uint64 `json:"size"`
	IsFreeleech     bool   `json:"isFreeleech"`
	TvMazeID        string `json:"tvmazeID,omitempty"`
	TvMazeEpisodeID string `json:"tvmazeEpisodeID,omitempty"`
}

type DownloadDeleteRequestInput struct {
	Reason string `json:"reason"`
}

type DownloadEventRecord struct {
	ID                uint64  `json:"id"`
	UserID            *uint64 `json:"userID,omitempty"`
	Username          string  `json:"username"`
	Fid               string  `json:"fid"`
	Filename          string  `json:"filename"`
	TvMazeID          string  `json:"tvmazeID,omitempty"`
	TvMazeEpisodeID   string  `json:"tvmazeEpisodeID,omitempty"`
	CategoryID        int     `json:"categoryID"`
	TorrentSize       uint64  `json:"torrentSize"`
	IsFreeleech       bool    `json:"isFreeleech"`
	QbtState          string  `json:"qbtState,omitempty"`
	ProgressPercent   float64 `json:"progressPercent"`
	CreatedAt         string  `json:"createdAt"`
	DeletedAt         *string `json:"deletedAt,omitempty"`
	DeletedByUsername *string `json:"deletedByUsername,omitempty"`
	HasPendingDelete  bool    `json:"hasPendingDeleteRequest"`
	HasHitAndRun      bool    `json:"hasHitAndRun"`
	CompletedAt       *string `json:"completedAt,omitempty"`
	SafeToDeleteAt    *string `json:"safeToDeleteAt,omitempty"`
	QbtHash           string  `json:"-"`
}

type DownloadDeleteRequestRecord struct {
	ID                  uint64  `json:"id"`
	DownloadEventID     uint64  `json:"downloadEventID"`
	RequestedByUsername string  `json:"requestedByUsername"`
	Status              string  `json:"status"`
	Reason              string  `json:"reason"`
	ApprovedByUsername  string  `json:"approvedByUsername,omitempty"`
	CreatedAt           string  `json:"createdAt"`
	ApprovedAt          string  `json:"approvedAt,omitempty"`
	SafeToDeleteAt      *string `json:"safeToDeleteAt,omitempty"`
	AutoDeleteAt        *string `json:"autoDeleteAt,omitempty"`
	DownloadFilename    string  `json:"downloadFilename,omitempty"`
	DownloadFid         string  `json:"downloadFid,omitempty"`
	DownloadSize        uint64  `json:"downloadSize"`
	DownloadIsFreeleech bool    `json:"downloadIsFreeleech"`
}
