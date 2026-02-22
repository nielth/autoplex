package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrTvMazeNotFound = errors.New("tvmaze resource not found")

const tvMazeBaseURL = "https://api.tvmaze.com"

type TvMazeImage struct {
	Medium   string `json:"medium"`
	Original string `json:"original"`
}

type TvMazeCountry struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Timezone string `json:"timezone"`
}

type TvMazeNetwork struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Country      *TvMazeCountry `json:"country"`
	OfficialSite string         `json:"officialSite"`
}

type TvMazeWebChannel struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Country      *TvMazeCountry `json:"country"`
	OfficialSite string         `json:"officialSite"`
}

type TvMazeSchedule struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

type TvMazeRating struct {
	Average *float64 `json:"average"`
}

type TvMazeLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type TvMazeShowLinks struct {
	Self            *TvMazeLink `json:"self"`
	PreviousEpisode *TvMazeLink `json:"previousepisode"`
	NextEpisode     *TvMazeLink `json:"nextepisode"`
}

type TvMazeShow struct {
	ID             int64             `json:"id"`
	URL            string            `json:"url"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Language       string            `json:"language"`
	Genres         []string          `json:"genres"`
	Status         string            `json:"status"`
	Runtime        *int              `json:"runtime"`
	AverageRuntime *int              `json:"averageRuntime"`
	Premiered      string            `json:"premiered"`
	Ended          *string           `json:"ended"`
	OfficialSite   string            `json:"officialSite"`
	Schedule       TvMazeSchedule    `json:"schedule"`
	Rating         TvMazeRating      `json:"rating"`
	Weight         int               `json:"weight"`
	Network        *TvMazeNetwork    `json:"network"`
	WebChannel     *TvMazeWebChannel `json:"webChannel"`
	Image          *TvMazeImage      `json:"image"`
	Summary        string            `json:"summary"`
	Updated        int64             `json:"updated"`
	Links          TvMazeShowLinks   `json:"_links"`
}

type TvMazeSearchResult struct {
	Score float64    `json:"score"`
	Show  TvMazeShow `json:"show"`
}

type TvMazeSeason struct {
	ID           int64        `json:"id"`
	URL          string       `json:"url"`
	Number       int          `json:"number"`
	Name         string       `json:"name"`
	EpisodeOrder *int         `json:"episodeOrder"`
	PremiereDate string       `json:"premiereDate"`
	EndDate      string       `json:"endDate"`
	Image        *TvMazeImage `json:"image"`
	Summary      *string      `json:"summary"`
}

type TvMazeEpisode struct {
	ID       int64        `json:"id"`
	URL      string       `json:"url"`
	Name     string       `json:"name"`
	Season   int          `json:"season"`
	Number   int          `json:"number"`
	Type     string       `json:"type"`
	Airdate  string       `json:"airdate"`
	Airtime  string       `json:"airtime"`
	Airstamp string       `json:"airstamp"`
	Runtime  *int         `json:"runtime"`
	Rating   TvMazeRating `json:"rating"`
	Image    *TvMazeImage `json:"image"`
	Summary  *string      `json:"summary"`
}

func TvMazeSearchShows(query string) ([]TvMazeSearchResult, error) {
	clean := strings.TrimSpace(query)
	if clean == "" {
		return []TvMazeSearchResult{}, nil
	}

	path := fmt.Sprintf("/search/shows?q=%s", url.QueryEscape(clean))
	var results []TvMazeSearchResult
	if err := tvMazeGet(path, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func TvMazeGetShow(showID int64) (*TvMazeShow, error) {
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	var show TvMazeShow
	if err := tvMazeGet(fmt.Sprintf("/shows/%d", showID), &show); err != nil {
		return nil, err
	}

	return &show, nil
}

func TvMazeGetSeasons(showID int64) ([]TvMazeSeason, error) {
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	var seasons []TvMazeSeason
	if err := tvMazeGet(fmt.Sprintf("/shows/%d/seasons", showID), &seasons); err != nil {
		return nil, err
	}

	return seasons, nil
}

func TvMazeGetEpisodes(showID int64) ([]TvMazeEpisode, error) {
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	var episodes []TvMazeEpisode
	if err := tvMazeGet(fmt.Sprintf("/shows/%d/episodes", showID), &episodes); err != nil {
		return nil, err
	}

	return episodes, nil
}

func TvMazeGetEpisode(episodeID int64) (*TvMazeEpisode, error) {
	if episodeID <= 0 {
		return nil, fmt.Errorf("episode id is required")
	}

	var episode TvMazeEpisode
	if err := tvMazeGet(fmt.Sprintf("/episodes/%d", episodeID), &episode); err != nil {
		return nil, err
	}

	return &episode, nil
}

func ParseTvMazeAirstamp(airstamp string) (time.Time, error) {
	clean := strings.TrimSpace(airstamp)
	if clean == "" {
		return time.Time{}, fmt.Errorf("airstamp is required")
	}

	parsed, err := time.Parse(time.RFC3339, clean)
	if err != nil {
		return time.Time{}, err
	}

	return parsed.UTC(), nil
}

func tvMazeGet(path string, result any) error {
	url := tvMazeBaseURL + path

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrTvMazeNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tvmaze request failed with status %d: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return err
	}

	return nil
}
