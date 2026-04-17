package services

import (
	"api/models"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const tlRssPollInterval = 3 * time.Minute

var (
	tlRssWorkerOnce sync.Once
	tlRssFidPattern = regexp.MustCompile(`/download/(\d+)/`)
	tlRssSeenFids   sync.Map
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

	match, quality := matchTlRssItem(title, subs)
	if match == nil {
		return
	}

	log.Printf("tl rss match: show=%q quality=%s fid=%s title=%q", match.showName, quality, fid, title)

	already, err := IsFidAlreadyDownloaded(fid)
	if err != nil {
		log.Printf("tl rss dup check failed fid=%s: %v", fid, err)
		return
	}
	tlRssSeenFids.Store(fid, struct{}{})
	if already {
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

func matchTlRssItem(title string, subs []tlRssSubscriptionMatch) (*tlRssSubscriptionMatch, string) {
	tokens := tokenizeTitle(title)
	if len(tokens) == 0 {
		return nil, ""
	}

	for i := range subs {
		sub := &subs[i]
		if !tokensStartWithShow(tokens, sub.showTokens) {
			continue
		}
		for normalized := range sub.qualities {
			if hasQualityToken(title, normalized) {
				return sub, normalized
			}
		}
	}
	return nil, ""
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
