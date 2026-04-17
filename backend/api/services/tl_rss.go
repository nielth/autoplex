package services

import (
	"api/models"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const tlRssPollInterval = 3 * time.Minute

var (
	tlRssWorkerOnce         sync.Once
	tlRssFidPattern         = regexp.MustCompile(`/download/(\d+)/`)
	tlRssEpisodeTokenRegex  = regexp.MustCompile(`^s0*([1-9][0-9]?)e0*([1-9][0-9]{0,2})$`)
	tlRssSeenFids           sync.Map
)

type tlRssFeed struct {
	XMLName xml.Name     `xml:"rss"`
	Channel tlRssChannel `xml:"channel"`
}

type tlRssChannel struct {
	Items []tlRssItem `xml:"item"`
}

type tlRssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	GUID    string `xml:"guid"`
}

type tlRssSubscriptionMatch struct {
	userID       uint64
	username     string
	tvmazeShowID int64
	showName     string
	showTokens   []string
	qualities    map[string]struct{}
}

func StartTlRssWorker() {
	tlRssWorkerOnce.Do(func() {
		if strings.TrimSpace(os.Getenv("TL_RSS_URL")) == "" {
			log.Printf("tl rss worker disabled: TL_RSS_URL not set")
			return
		}
		go runTlRssWorker()
	})
}

func runTlRssWorker() {
	ticker := time.NewTicker(tlRssPollInterval)
	defer ticker.Stop()

	runTlRssCycle()
	for range ticker.C {
		runTlRssCycle()
	}
}

func runTlRssCycle() {
	url := strings.TrimSpace(os.Getenv("TL_RSS_URL"))
	if url == "" {
		return
	}

	items, err := fetchTlRssFeed(url)
	if err != nil {
		log.Printf("tl rss fetch failed: %v", err)
		return
	}

	subs, err := loadTlRssSubscriptionIndex()
	if err != nil {
		log.Printf("tl rss subscription load failed: %v", err)
		return
	}
	if len(subs) == 0 {
		return
	}

	for _, item := range items {
		processTlRssItem(item, subs)
	}
}

func fetchTlRssFeed(url string) ([]tlRssItem, error) {
	body, err := tlGetRequest(url, nil)
	if err != nil {
		return nil, err
	}

	var feed tlRssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}
	return feed.Channel.Items, nil
}

func loadTlRssSubscriptionIndex() ([]tlRssSubscriptionMatch, error) {
	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT s.user_id, s.username, s.tvmaze_show_id, COALESCE(s.show_name, ''), q.preferred_quality
		FROM tv_show_subscriptions s
		INNER JOIN tv_show_auto_install_qualities q
		  ON q.user_id = s.user_id
		  AND q.tvmaze_show_id = s.tvmaze_show_id
		  AND q.enabled = 1
		WHERE s.enabled = 1`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexByKey := map[string]*tlRssSubscriptionMatch{}
	var ordered []*tlRssSubscriptionMatch

	for rows.Next() {
		var r tlRssSubscriptionMatch
		var quality string
		if err := rows.Scan(&r.userID, &r.username, &r.tvmazeShowID, &r.showName, &quality); err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%d-%d", r.userID, r.tvmazeShowID)
		existing, ok := indexByKey[key]
		if !ok {
			r.showTokens = tokenizeTitle(r.showName)
			r.qualities = map[string]struct{}{}
			indexByKey[key] = &r
			ordered = append(ordered, &r)
			existing = &r
		}
		existing.qualities[NormalizeQualityPreference(quality)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]tlRssSubscriptionMatch, 0, len(ordered))
	for _, sub := range ordered {
		if len(sub.showTokens) == 0 || len(sub.qualities) == 0 {
			continue
		}
		out = append(out, *sub)
	}
	return out, nil
}

func processTlRssItem(item tlRssItem, subs []tlRssSubscriptionMatch) {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		return
	}

	link := strings.TrimSpace(item.Link)
	fid := extractTlFidFromLink(link)
	if fid == "" {
		return
	}

	if _, seen := tlRssSeenFids.Load(fid); seen {
		return
	}

	match, quality, season, episode := matchTlRssItem(title, subs)
	if match == nil {
		return
	}

	log.Printf("tl rss match: show=%q s%02de%02d quality=%s fid=%s title=%q", match.showName, season, episode, quality, fid, title)

	already, err := IsFidAlreadyDownloaded(fid)
	if err != nil {
		log.Printf("tl rss dup check failed fid=%s: %v", fid, err)
		return
	}
	tlRssSeenFids.Store(fid, struct{}{})
	if already {
		return
	}

	episodeOwned, err := isEpisodeAlreadyDownloaded(match.tvmazeShowID, season, episode, quality)
	if err != nil {
		log.Printf("tl rss episode ownership check failed show=%d s%02de%02d: %v", match.tvmazeShowID, season, episode, err)
		return
	}
	if episodeOwned {
		log.Printf("tl rss skip: show=%q s%02de%02d already downloaded", match.showName, season, episode)
		return
	}

	filename := deriveTlRssFilename(link, title)
	data := models.DownloadData{
		Fid:        fid,
		Filename:   filename,
		CategoryID: tvDefaultCategoryID,
		TvMazeID:   fmt.Sprintf("%d", match.tvmazeShowID),
	}

	qbtHash, err := TlDownloadRequest(data)
	if err != nil {
		log.Printf("tl rss download failed title=%q fid=%s: %v", title, fid, err)
		return
	}

	if logErr := LogDownloadEvent(match.username, data, qbtHash, true, "", "", "tv-rss-worker"); logErr != nil {
		log.Printf("tl rss failed to log event fid=%s: %v", fid, logErr)
	}
	ScheduleAutoPlexScanForDownload(qbtHash, data.CategoryID)
	log.Printf("tl rss dispatched download show=%q fid=%s quality=%s title=%q", match.showName, fid, quality, title)
}

func extractTlFidFromLink(link string) string {
	match := tlRssFidPattern.FindStringSubmatch(link)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func deriveTlRssFilename(link string, title string) string {
	if idx := strings.LastIndex(link, "/"); idx > -1 && idx < len(link)-1 {
		if name := link[idx+1:]; name != "" {
			return name
		}
	}
	return strings.ReplaceAll(title, " ", ".")
}

func matchTlRssItem(title string, subs []tlRssSubscriptionMatch) (*tlRssSubscriptionMatch, string, int, int) {
	tokens := tokenizeTitle(title)
	if len(tokens) == 0 {
		return nil, "", 0, 0
	}

	season, episode, hasEpisode := extractSingleEpisodeToken(tokens)
	if !hasEpisode {
		return nil, "", 0, 0
	}

	for i := range subs {
		sub := &subs[i]
		if !tokensStartWithShow(tokens, sub.showTokens) {
			continue
		}
		for normalized := range sub.qualities {
			if hasQualityToken(title, normalized) {
				return sub, normalized, season, episode
			}
		}
	}
	return nil, "", 0, 0
}

func extractSingleEpisodeToken(tokens []string) (int, int, bool) {
	for _, token := range tokens {
		m := tlRssEpisodeTokenRegex.FindStringSubmatch(token)
		if len(m) != 3 {
			continue
		}
		season, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		episode, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		return season, episode, true
	}
	return 0, 0, false
}

func tokensStartWithShow(titleTokens []string, showTokens []string) bool {
	if len(showTokens) == 0 || len(titleTokens) <= len(showTokens) {
		return false
	}
	for i, token := range showTokens {
		if titleTokens[i] != token {
			return false
		}
	}
	return true
}

func isEpisodeAlreadyDownloaded(tvmazeShowID int64, season int, episode int, quality string) (bool, error) {
	db, err := dbConn()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err = db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tv_episode_jobs
			WHERE tvmaze_show_id = ?
			  AND season_number = ?
			  AND episode_number = ?
			  AND preferred_quality = ?
			  AND status = 'downloaded'
		)`,
		tvmazeShowID,
		season,
		episode,
		NormalizeQualityPreference(quality),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
