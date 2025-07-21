package utils

type TorrentDisplay struct {
	Fid                string
	Filename           string
	Name               string
	AddedTimestamp     string
	CategoryID         int
	Size               int64
	FormattedSize      string // New
	IsFreeLeech        bool   // New
	Completed          int
	Seeders            int
	Leechers           int
	NumComments        int
	Tags               any
	New                bool
	ImdbID             string
	Rating             float64
	Genres             string
	TvmazeID           string
	IgdbID             any
	AnimeID            string
	DownloadMultiplier int
	CommentsDisabled   int
	Uploader           string
}

type TorrentListStruct struct {
	NumFound    int              `json:"numFound,omitempty"`
	TorrentList []TorrentDisplay `json:"torrentList,omitempty"`
	Order       string           `json:"order,omitempty"`
	OrderBy     string           `json:"orderBy,omitempty"`
	Page        int              `json:"page,omitempty"`
	PerPage     int              `json:"perPage,omitempty"`
}
