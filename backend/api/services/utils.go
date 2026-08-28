package services

import (
	"api/models"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var tlHttpClient = buildTlHttpClient()

func buildTlHttpClient() *http.Client {
	proxyURL := strings.TrimSpace(os.Getenv("TL_PROXY_URL"))
	if proxyURL == "" {
		return &http.Client{}
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("invalid TL_PROXY_URL %q, falling back to direct: %v", proxyURL, err)
		return &http.Client{}
	}

	log.Printf("routing TL requests through proxy %s", parsed.Redacted())
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsed),
		},
	}
}

func fetchAndParseXML(url string, result any, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		errCh <- fmt.Errorf("failed to fetch URL %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errCh <- fmt.Errorf("failed to read response body from %s: %v", url, err)
		return
	}

	if err := xml.Unmarshal(body, result); err != nil {
		errCh <- fmt.Errorf("failed to parse XML from %s: %v", url, err)
		return
	}
}

func readCookie() (map[string]string, error) {
	raw := os.Getenv("TL_COOKIE_JSON")
	if raw == "" {
		return nil, fmt.Errorf("TL_COOKIE_JSON environment variable is not set")
	}

	var cookieData map[string]string
	if err := json.Unmarshal([]byte(raw), &cookieData); err != nil {
		fmt.Printf("Error parsing TL_COOKIE_JSON: %v\n", err)
		return nil, fmt.Errorf("Error parsing TL_COOKIE_JSON")
	}

	if len(cookieData) == 0 {
		return nil, fmt.Errorf("TL_COOKIE_JSON is empty")
	}

	return cookieData, nil
}

func tlGetRequest(url string, ua *string) ([]byte, error) {
	cookieData, err := readCookie()

	if err != nil {
		fmt.Printf("Error reading Cookie: %v\n", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, fmt.Errorf("Error creating request")
	}

	for key, value := range cookieData {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	if ua != nil {
		req.Header.Add("User-Agent", *ua)
	}

	resp, err := tlHttpClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, fmt.Errorf("Error sending request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, fmt.Errorf("Error reading request body")
	}

	return body, nil
}

func tlPostFormRequest(rawURL string, values url.Values, ua *string) ([]byte, error) {
	cookieData, err := readCookie()
	if err != nil {
		fmt.Printf("Error reading Cookie: %v\n", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, fmt.Errorf("Error creating request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range cookieData {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	if ua != nil {
		req.Header.Add("User-Agent", *ua)
	}

	resp, err := tlHttpClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, fmt.Errorf("Error sending request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, fmt.Errorf("Error reading request body")
	}

	return body, nil
}

func TlSearchRequest(search string, page string) (map[string]any, error) {
	url := "https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,34,26,32,27,44,36/query/" + search + "/orderby/seeders/order/desc/page/" + page

	body, err := tlGetRequest(url, nil)

	var respJson map[string]any
	if err := json.Unmarshal(body, &respJson); err != nil {
		fmt.Printf("Cannot read JSON from request: %v\n", err)
		return nil, fmt.Errorf("Cannot read JSON from request")
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if torrents, exists := respJson["torrentList"].([]any); exists {
		filteredTorrents := []any{}
		for _, t := range torrents {
			torrent := t.(map[string]any)
			tags, tagExists := torrent["tags"].([]any)
			if tagExists {
				containsDolbyVision := false
				for _, tag := range tags {
					tagStr, ok := tag.(string)
					if ok && strings.Contains(strings.ToLower(tagStr), "dolby vision") {
						containsDolbyVision = true
						break
					}
				}
				if !containsDolbyVision {
					filteredTorrents = append(filteredTorrents, torrent)
				}
			} else {
				// Add torrents without tags
				filteredTorrents = append(filteredTorrents, torrent)
			}
		}
		respJson["torrentList"] = filteredTorrents
	}

	return respJson, nil
}

func TlDownloadRequest(data models.DownloadData, sequential bool) (string, error) {
	url := "https://www.torrentleech.org/download/" + data.Fid + "/" + data.Filename
	ua := "U_AGENT" // Without this, TL for some reason breaks

	torrent_data, err := tlGetRequest(url, &ua)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	category, ok := ResolveQbtCategory(data.CategoryID)
	if !ok {
		return "", fmt.Errorf("unsupported category id %d", data.CategoryID)
	}

	qbtHash, err := QbtDownload(&torrent_data, category, data.Fid, sequential, data.Filename)

	if err != nil {
		return "", err
	}

	return qbtHash, nil
}

type TlSeriesTorrent struct {
	Fid            string   `json:"fid"`
	Filename       string   `json:"filename"`
	CategoryID     int      `json:"categoryID"`
	Size           uint64   `json:"size"`
	Seeders        int      `json:"seeders"`
	Leechers       int      `json:"leechers"`
	Name           string   `json:"name"`
	Tags           []string `json:"tags"`
	AddedTimestamp int64    `json:"addedTimestamp"`
	TvMazeID       string   `json:"tvmazeID"`
}

var validQualityPreferences = []string{"1080", "2160"}

// Dynamic range preferences for 2160p auto-install:
//
//	any - take the best 2160p release regardless of dynamic range
//	dv  - prefer Dolby Vision, fall back to plain HDR when no DV release exists
//	hdr - only take HDR releases that are not also Dolby Vision
var validDynamicRangePreferences = []string{"any", "dv", "hdr"}

func NormalizeQualityPreference(value string) string {
	clean := strings.TrimSpace(strings.ToLower(value))
	if clean == "2160p" {
		clean = "2160"
	}
	if clean == "1080p" {
		clean = "1080"
	}
	if slices.Contains(validQualityPreferences, clean) {
		return clean
	}
	return "1080"
}

func NormalizeDynamicRangePreference(value string) string {
	clean := strings.TrimSpace(strings.ToLower(value))
	switch clean {
	case "dovi", "dolbyvision", "dolby vision", "dv":
		clean = "dv"
	case "hdr10", "hdr10+", "hdr":
		clean = "hdr"
	}
	if slices.Contains(validDynamicRangePreferences, clean) {
		return clean
	}
	return "any"
}

// EffectiveDynamicRange only lets the preference through for 2160p; 1080p
// releases are practically never tagged DV/HDR, so a stale preference there
// would filter every candidate away.
func EffectiveDynamicRange(quality string, dynamicRange string) string {
	if NormalizeQualityPreference(quality) != "2160" {
		return "any"
	}
	return NormalizeDynamicRangePreference(dynamicRange)
}

func TlSeriesSearchByTvMaze(tvmazeEpisodeID int64, tvmazeID int64) ([]TlSeriesTorrent, error) {
	if tvmazeEpisodeID <= 0 {
		return nil, fmt.Errorf("tvmaze episode id is required")
	}
	if tvmazeID <= 0 {
		return nil, fmt.Errorf("tvmaze id is required")
	}

	payload := url.Values{}
	payload.Set("tvmazeEpisodeID", fmt.Sprintf("%d", tvmazeEpisodeID))
	payload.Set("tvmazeID", fmt.Sprintf("%d", tvmazeID))

	body, err := tlPostFormRequest("https://www.torrentleech.org/torrents/series/torrent", payload, nil)
	if err != nil {
		return nil, err
	}

	var torrentList []TlSeriesTorrent
	if err := json.Unmarshal(body, &torrentList); err != nil {
		fmt.Printf("Cannot read JSON from tvmaze series request: %v\n", err)
		return nil, fmt.Errorf("cannot read torrent list response")
	}

	return torrentList, nil
}

func TlSeriesBoxsetSearchByTvMaze(tvmazeSeriesID int64, tvmazeID int64) ([]TlSeriesTorrent, error) {
	if tvmazeSeriesID <= 0 {
		return nil, fmt.Errorf("tvmaze series id is required")
	}
	if tvmazeID <= 0 {
		return nil, fmt.Errorf("tvmaze id is required")
	}

	payload := url.Values{}
	payload.Set("tvmazeSeriesID", fmt.Sprintf("%d", tvmazeSeriesID))
	payload.Set("tvmazeID", fmt.Sprintf("%d", tvmazeID))

	body, err := tlPostFormRequest("https://www.torrentleech.org/torrents/series/boxset", payload, nil)
	if err != nil {
		return nil, err
	}

	var torrentList []TlSeriesTorrent
	if err := json.Unmarshal(body, &torrentList); err != nil {
		fmt.Printf("Cannot read JSON from tvmaze boxset request: %v\n", err)
		return nil, fmt.Errorf("cannot read boxset response")
	}

	return torrentList, nil
}

func SelectBestTorrentByQuality(torrents []TlSeriesTorrent, quality string, dynamicRange string) *TlSeriesTorrent {
	if len(torrents) == 0 {
		return nil
	}

	preferredQuality := NormalizeQualityPreference(quality)
	var candidates []TlSeriesTorrent
	for _, torrent := range torrents {
		if !IsMovieOrTVCategory(torrent.CategoryID) {
			continue
		}
		if hasQualityToken(torrent.Name, preferredQuality) || hasQualityToken(torrent.Filename, preferredQuality) {
			candidates = append(candidates, torrent)
		}
	}

	return pickBestByDynamicRange(candidates, EffectiveDynamicRange(preferredQuality, dynamicRange))
}

func SelectBestBoxsetTorrentByQuality(torrents []TlSeriesTorrent, showName string, seasonNumber int, quality string, dynamicRange string) *TlSeriesTorrent {
	if len(torrents) == 0 {
		return nil
	}

	cleanShowName := strings.TrimSpace(showName)
	if cleanShowName == "" {
		return nil
	}
	if seasonNumber <= 0 {
		return nil
	}

	preferredQuality := NormalizeQualityPreference(quality)
	candidates := make([]TlSeriesTorrent, 0, len(torrents))
	for _, torrent := range torrents {
		if !IsMovieOrTVCategory(torrent.CategoryID) {
			continue
		}

		if boxsetNameMatchesShowSeasonQuality(torrent.Name, cleanShowName, seasonNumber, preferredQuality) ||
			boxsetNameMatchesShowSeasonQuality(torrent.Filename, cleanShowName, seasonNumber, preferredQuality) {
			candidates = append(candidates, torrent)
		}
	}

	return pickBestByDynamicRange(candidates, EffectiveDynamicRange(preferredQuality, dynamicRange))
}

// pickBestByDynamicRange narrows quality-matched candidates to the wanted
// dynamic range before picking the best one: "hdr" keeps HDR releases that are
// not also Dolby Vision, "dv" prefers Dolby Vision and falls back to those same
// plain HDR releases when no DV release exists, and "any" keeps every candidate.
func pickBestByDynamicRange(candidates []TlSeriesTorrent, dynamicRange string) *TlSeriesTorrent {
	switch dynamicRange {
	case "dv":
		if best := pickBestTorrent(filterTorrents(candidates, torrentHasDolbyVision)); best != nil {
			return best
		}
		return pickBestTorrent(filterTorrents(candidates, torrentHasHdrWithoutDolbyVision))
	case "hdr":
		return pickBestTorrent(filterTorrents(candidates, torrentHasHdrWithoutDolbyVision))
	default:
		return pickBestTorrent(candidates)
	}
}

func filterTorrents(torrents []TlSeriesTorrent, keep func(TlSeriesTorrent) bool) []TlSeriesTorrent {
	filtered := make([]TlSeriesTorrent, 0, len(torrents))
	for _, torrent := range torrents {
		if keep(torrent) {
			filtered = append(filtered, torrent)
		}
	}
	return filtered
}

// pickBestTorrent prefers the most seeded release, newest wins a tie.
func pickBestTorrent(candidates []TlSeriesTorrent) *TlSeriesTorrent {
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	for _, torrent := range candidates[1:] {
		if torrent.Seeders > best.Seeders {
			best = torrent
			continue
		}

		if torrent.Seeders == best.Seeders && torrent.AddedTimestamp > best.AddedTimestamp {
			best = torrent
		}
	}

	return &best
}

func torrentHasDolbyVision(torrent TlSeriesTorrent) bool {
	return HasDolbyVisionToken(torrent.Name) || HasDolbyVisionToken(torrent.Filename)
}

func torrentHasHdrWithoutDolbyVision(torrent TlSeriesTorrent) bool {
	if torrentHasDolbyVision(torrent) {
		return false
	}
	return HasHdrToken(torrent.Name) || HasHdrToken(torrent.Filename)
}

// HasDolbyVisionToken reports whether a release name is tagged Dolby Vision,
// e.g. "...DDP5.1.Atmos.DV.HDR.H.265-FLUX" or "...DoVi.HDR10...".
func HasDolbyVisionToken(value string) bool {
	tokens := tokenizeTitle(value)
	for index, token := range tokens {
		switch token {
		case "dv", "dovi", "dolbyvision":
			return true
		case "dolby":
			if index+1 < len(tokens) && tokens[index+1] == "vision" {
				return true
			}
		}
	}
	return false
}

// HasHdrToken reports whether a release name is tagged HDR (HDR, HDR10,
// HDR10+). DV releases usually carry an HDR tag too, so callers wanting plain
// HDR must rule out Dolby Vision separately.
func HasHdrToken(value string) bool {
	for _, token := range tokenizeTitle(value) {
		if strings.HasPrefix(token, "hdr") {
			return true
		}
	}
	return false
}

func boxsetNameMatchesShowSeasonQuality(value string, showName string, seasonNumber int, quality string) bool {
	tokens := tokenizeTitle(value)
	showTokens := tokenizeTitle(showName)
	if len(tokens) == 0 || len(showTokens) == 0 {
		return false
	}
	if len(tokens) <= len(showTokens) {
		return false
	}

	// Rule: torrent name starts with show name.
	for index, token := range showTokens {
		if tokens[index] != token {
			return false
		}
	}

	position := len(showTokens)
	if position < len(tokens) && isYearToken(tokens[position]) {
		position++
	}
	if position >= len(tokens) {
		return false
	}

	seasonTokenShort := fmt.Sprintf("s%d", seasonNumber)
	seasonTokenPadded := fmt.Sprintf("s%02d", seasonNumber)
	switch tokens[position] {
	case seasonTokenShort, seasonTokenPadded:
		position++
	case "season":
		if position+1 >= len(tokens) {
			return false
		}
		next := tokens[position+1]
		if next != fmt.Sprintf("%d", seasonNumber) && next != fmt.Sprintf("%02d", seasonNumber) {
			return false
		}
		position += 2
	default:
		return false
	}

	seasonNumbers := extractSeasonNumbers(tokens)
	if len(seasonNumbers) == 0 {
		return false
	}
	for _, foundSeason := range seasonNumbers {
		if foundSeason != seasonNumber {
			return false
		}
	}

	qualityToken := NormalizeQualityPreference(quality)
	for _, token := range tokens[position:] {
		if token == qualityToken || token == qualityToken+"p" {
			return true
		}
		if qualityToken == "2160" && (token == "uhd" || strings.Contains(token, "4k")) {
			return true
		}
	}

	return false
}

func tokenizeTitle(value string) []string {
	clean := strings.ToLower(strings.TrimSpace(value))
	if clean == "" {
		return nil
	}

	// Drop apostrophes so possessives collapse the way scene releases name them:
	// "Widow's Bay" -> "widows bay" matches "Widows.Bay...", not "widow s bay".
	clean = strings.NewReplacer("'", "", "’", "", "`", "").Replace(clean)

	return regexp.MustCompile(`[a-z0-9]+`).FindAllString(clean, -1)
}

func isYearToken(value string) bool {
	if len(value) != 4 {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	year := value[0:4]
	return year >= "1900" && year <= "2100"
}

var seasonTokenPattern = regexp.MustCompile(`^s0*([1-9][0-9]?)(?:e[0-9]+)?$`)

func extractSeasonNumbers(tokens []string) []int {
	seasons := make([]int, 0)
	for index := 0; index < len(tokens); index++ {
		token := tokens[index]

		if token == "season" {
			if index+1 >= len(tokens) {
				continue
			}
			seasonValue, err := strconv.Atoi(tokens[index+1])
			if err != nil || seasonValue <= 0 || seasonValue > 99 {
				continue
			}
			seasons = append(seasons, seasonValue)
			index++
			continue
		}

		matches := seasonTokenPattern.FindStringSubmatch(token)
		if len(matches) != 2 {
			continue
		}
		seasonValue, err := strconv.Atoi(matches[1])
		if err != nil || seasonValue <= 0 || seasonValue > 99 {
			continue
		}
		seasons = append(seasons, seasonValue)
	}

	return seasons
}

func hasQualityToken(value string, quality string) bool {
	clean := strings.ToLower(strings.TrimSpace(value))
	if clean == "" {
		return false
	}

	target := NormalizeQualityPreference(quality)
	pattern := fmt.Sprintf(`\b%s(p)?\b`, regexp.QuoteMeta(target))
	matched, err := regexp.MatchString(pattern, clean)
	if err != nil {
		return strings.Contains(clean, target)
	}

	return matched
}

func seasonTokenMatches(value string, seasonNumber int) bool {
	if seasonNumber <= 0 {
		return false
	}

	clean := strings.ToLower(strings.TrimSpace(value))
	if clean == "" {
		return false
	}

	patterns := []string{
		fmt.Sprintf(`\bs%02d\b`, seasonNumber),
		fmt.Sprintf(`\bs%d\b`, seasonNumber),
		fmt.Sprintf(`\bseason[\s._-]*%d\b`, seasonNumber),
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, clean)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

// deriveTorrentFilename picks a human-readable name for the download record.
// Some TL uploads carry a useless placeholder filename ("torrent.torrent") or
// none at all; fall back to the release name so /downloads shows the real
// title instead of "torrent.torrent".
func deriveTorrentFilename(filename string, name string) string {
	clean := strings.TrimSpace(filename)
	if clean == "" || strings.EqualFold(clean, "torrent.torrent") {
		if release := strings.TrimSpace(name); release != "" {
			return strings.ReplaceAll(release, " ", ".") + ".torrent"
		}
	}
	return clean
}

func ConvertTlSeriesTorrentToDownloadData(torrent TlSeriesTorrent, tvmazeID int64, tvmazeEpisodeID int64) models.DownloadData {
	tags := make(map[string]struct{}, len(torrent.Tags))
	for _, tag := range torrent.Tags {
		tags[strings.ToUpper(strings.TrimSpace(tag))] = struct{}{}
	}

	_, isFreeleech := tags["FREELEECH"]

	return models.DownloadData{
		Fid:             strings.TrimSpace(torrent.Fid),
		Filename:        deriveTorrentFilename(torrent.Filename, torrent.Name),
		CategoryID:      torrent.CategoryID,
		Size:            torrent.Size,
		IsFreeleech:     isFreeleech,
		TvMazeID:        fmt.Sprintf("%d", tvmazeID),
		TvMazeEpisodeID: fmt.Sprintf("%d", tvmazeEpisodeID),
	}
}

func NextEpisodeRetryTime(airstamp time.Time, now time.Time, airtimeKnown bool) (time.Time, bool) {
	if !airtimeKnown {
		rapidCheckEnd := airstamp.Add(48 * time.Hour)
		dailyCheckEnd := rapidCheckEnd.Add(7 * 24 * time.Hour)

		if now.Before(airstamp) {
			return airstamp, true
		}
		if now.Before(rapidCheckEnd) {
			return now.Add(30 * time.Minute), true
		}
		if now.Before(dailyCheckEnd) {
			return now.Add(24 * time.Hour), true
		}
		return time.Time{}, false
	}

	releaseCheckStart := airstamp.Add(30 * time.Minute)
	rapidCheckEnd := releaseCheckStart.Add(2 * time.Hour)
	dailyCheckEnd := rapidCheckEnd.Add(7 * 24 * time.Hour)

	if now.Before(releaseCheckStart) {
		return releaseCheckStart, true
	}
	if now.Before(rapidCheckEnd) {
		return now.Add(10 * time.Minute), true
	}
	if now.Before(dailyCheckEnd) {
		return now.Add(24 * time.Hour), true
	}

	return time.Time{}, false
}

func DiskUsage() (map[string]uint64, error) {
	var stat syscall.Statfs_t

	path := os.Getenv("PATH_DISK")
	if path == "" {
		return nil, fmt.Errorf("PATH_DISK environment variable is not set")
	}

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("Error reading PATH: %v", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return map[string]uint64{"total": total, "free": free, "used": used}, nil
}
