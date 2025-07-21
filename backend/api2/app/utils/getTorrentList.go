package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
)

func GetTorrentList(torrentSearch string) (*TorrentListStruct, error) {
	var torrentList TorrentListStruct

	tlCookie := os.Getenv("TL_COOKIE")

	url := "https://www.torrentleech.org/torrents/browse/list/query/%s/orderby/seeders/order/desc"
	urlSearch := fmt.Sprintf(url, torrentSearch)

	req, err := http.NewRequest("GET", urlSearch, nil)
	if err != nil {
		return nil, err
	}

	if tlCookie == "" {
		return nil, fmt.Errorf("Could not read TL_COOKIE from .env, system error")

	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", tlCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // Important to close the response body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &torrentList)
	if err != nil {
		return nil, err
	}

	for i := range torrentList.TorrentList {
		torrentList.TorrentList[i].FormattedSize = ByteCountIEC(torrentList.TorrentList[i].Size)

		var tags []string
		switch t := torrentList.TorrentList[i].Tags.(type) {
		case []string:
			tags = t
		case []any:
			for _, tag := range t {
				if s, ok := tag.(string); ok {
					tags = append(tags, s)
				}
			}
		}
		torrentList.TorrentList[i].IsFreeLeech = slices.Contains(tags, "FREELEECH")
	}

	return &torrentList, nil
}
