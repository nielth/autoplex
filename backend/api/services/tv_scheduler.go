package services

import (
	"api/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	tvDefaultCategoryID       = 32
	tvEpisodeWorkerInterval   = 1 * time.Minute
	tvEpisodeWorkerBatchLimit = 25
)

var (
	ErrSeasonIncomplete = errors.New("season is incomplete")
	ErrEpisodeNotAired  = errors.New("episode has not aired yet")
	ErrNoQualityMatch   = errors.New("no matching torrent found for selected quality")
)

type TvShowSubscriptionRecord struct {
	ID                  uint64 `json:"id"`
	TvMazeShowID        int64  `json:"tvmazeShowID"`
	ShowName            string `json:"showName"`
	PreferredQuality    string `json:"preferredQuality"`
	AutoInstallUpcoming bool   `json:"autoInstallUpcoming"`
	Enabled             bool   `json:"enabled"`
	UpdatedAt           string `json:"updatedAt"`
}

type TvEpisodeJobRecord struct {
	ID               uint64  `json:"id"`
	TvMazeShowID     int64   `json:"tvmazeShowID"`
	TvMazeEpisodeID  int64   `json:"tvmazeEpisodeID"`
	EpisodeName      string  `json:"episodeName"`
	SeasonNumber     int     `json:"seasonNumber"`
	EpisodeNumber    int     `json:"episodeNumber"`
	Airstamp         *string `json:"airstamp,omitempty"`
	PreferredQuality string  `json:"preferredQuality"`
	Status           string  `json:"status"`
	AttemptCount     uint64  `json:"attemptCount"`
	NextCheckAt      string  `json:"nextCheckAt"`
	LastCheckedAt    *string `json:"lastCheckedAt,omitempty"`
	LastError        string  `json:"lastError,omitempty"`
}

type TvShowInstallStatus struct {
	Subscription    *TvShowSubscriptionRecord `json:"subscription,omitempty"`
	Jobs            []TvEpisodeJobRecord      `json:"jobs"`
	PendingCount    int                       `json:"pendingCount"`
	DownloadedCount int                       `json:"downloadedCount"`
	FailedCount     int                       `json:"failedCount"`
}

type TvInstallQueueResult struct {
	Queued    int `json:"queued"`
	Skipped   int `json:"skipped"`
	Triggered int `json:"triggered"`
}

type tvEpisodeJobRow struct {
	ID               uint64
	Username         string
	TvMazeShowID     int64
	TvMazeEpisodeID  int64
	EpisodeName      string
	SeasonNumber     sql.NullInt64
	EpisodeNumber    sql.NullInt64
	Airstamp         sql.NullTime
	PreferredQuality string
	AttemptCount     uint64
}

var (
	tvEpisodeWorkerOnce sync.Once
	tvEpisodeWorkerMu   sync.Mutex
)

func StartTvEpisodeAutoInstallWorker() {
	tvEpisodeWorkerOnce.Do(func() {
		go runTvEpisodeAutoInstallWorker()
	})
}

func runTvEpisodeAutoInstallWorker() {
	ticker := time.NewTicker(tvEpisodeWorkerInterval)
	defer ticker.Stop()

	if err := ProcessDueTvEpisodeJobs(tvEpisodeWorkerBatchLimit); err != nil {
		log.Printf("tv episode worker initial run failed: %v", err)
	}

	for range ticker.C {
		if err := ProcessDueTvEpisodeJobs(tvEpisodeWorkerBatchLimit); err != nil {
			log.Printf("tv episode worker run failed: %v", err)
		}
	}
}

func ProcessDueTvEpisodeJobs(limit int) error {
	if limit <= 0 {
		limit = tvEpisodeWorkerBatchLimit
	}

	tvEpisodeWorkerMu.Lock()
	defer tvEpisodeWorkerMu.Unlock()

	jobs, err := listDueEpisodeJobs(limit)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := processEpisodeJob(job); err != nil {
			log.Printf("failed processing tv episode job %d: %v", job.ID, err)
		}
	}

	return nil
}

func listDueEpisodeJobs(limit int) ([]tvEpisodeJobRow, error) {
	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT
			j.id,
			j.username,
			j.tvmaze_show_id,
			j.tvmaze_episode_id,
			COALESCE(j.episode_name, ''),
			j.season_number,
			j.episode_number,
			j.airstamp,
			j.preferred_quality,
			j.attempt_count
		FROM tv_episode_jobs j
		LEFT JOIN tv_show_subscriptions s
			ON s.id = j.subscription_id
		WHERE j.status IN ('pending', 'searching')
		  AND j.next_check_at <= UTC_TIMESTAMP()
		  AND (j.subscription_id IS NULL OR (s.enabled = 1 AND s.auto_install_upcoming = 1))
		ORDER BY j.next_check_at ASC, j.id ASC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]tvEpisodeJobRow, 0, limit)
	for rows.Next() {
		var row tvEpisodeJobRow
		if err := rows.Scan(
			&row.ID,
			&row.Username,
			&row.TvMazeShowID,
			&row.TvMazeEpisodeID,
			&row.EpisodeName,
			&row.SeasonNumber,
			&row.EpisodeNumber,
			&row.Airstamp,
			&row.PreferredQuality,
			&row.AttemptCount,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func processEpisodeJob(job tvEpisodeJobRow) error {
	torrents, err := TlSeriesSearchByTvMaze(job.TvMazeEpisodeID, job.TvMazeShowID)
	if err != nil {
		return markEpisodeJobRetry(job, fmt.Sprintf("torrent search failed: %v", err))
	}

	selected := SelectBestTorrentByQuality(torrents, job.PreferredQuality)
	if selected == nil {
		return markEpisodeJobRetry(job, fmt.Sprintf("no %sp torrent found yet", NormalizeQualityPreference(job.PreferredQuality)))
	}

	downloadData := ConvertTlSeriesTorrentToDownloadData(*selected, job.TvMazeShowID, job.TvMazeEpisodeID)
	if !IsMovieOrTVCategory(downloadData.CategoryID) {
		downloadData.CategoryID = tvDefaultCategoryID
	}

	isDownloaded, err := IsFidAlreadyDownloaded(downloadData.Fid)
	if err != nil {
		return markEpisodeJobRetry(job, fmt.Sprintf("failed duplicate check: %v", err))
	}

	var downloadEventID *uint64
	if isDownloaded {
		eventID, lookupErr := findLatestDownloadEventIDByFid(downloadData.Fid, job.Username)
		if lookupErr == nil {
			downloadEventID = eventID
		}
	} else {
		qbtHash, downloadErr := TlDownloadRequest(downloadData)
		if downloadErr != nil {
			return markEpisodeJobRetry(job, fmt.Sprintf("download request failed: %v", downloadErr))
		}

		if logErr := LogDownloadEvent(job.Username, downloadData, qbtHash, true, "", "", "tv-auto-install-worker"); logErr != nil {
			log.Printf("failed to write automated download event for job %d: %v", job.ID, logErr)
		}
		ScheduleAutoPlexScanForDownload(qbtHash, downloadData.CategoryID)

		eventID, lookupErr := findLatestDownloadEventIDByFid(downloadData.Fid, job.Username)
		if lookupErr == nil {
			downloadEventID = eventID
		}
	}

	return markEpisodeJobDownloaded(job.ID, downloadEventID)
}

func markEpisodeJobRetry(job tvEpisodeJobRow, reason string) error {
	now := time.Now().UTC()
	var nextCheck time.Time
	var hasNext bool

	if job.Airstamp.Valid {
		nextCheck, hasNext = NextEpisodeRetryTime(job.Airstamp.Time.UTC(), now)
	}

	if !hasNext {
		return markEpisodeJobFailed(job.ID, reason)
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`UPDATE tv_episode_jobs
		SET status = 'searching',
			attempt_count = attempt_count + 1,
			last_checked_at = UTC_TIMESTAMP(),
			next_check_at = ?,
			last_error = ?,
			updated_at = UTC_TIMESTAMP()
		WHERE id = ?`,
		nextCheck,
		nullableString(reason),
		job.ID,
	)
	return err
}

func markEpisodeJobFailed(jobID uint64, reason string) error {
	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`UPDATE tv_episode_jobs
		SET status = 'failed',
			attempt_count = attempt_count + 1,
			last_checked_at = UTC_TIMESTAMP(),
			last_error = ?,
			updated_at = UTC_TIMESTAMP()
		WHERE id = ?`,
		nullableString(reason),
		jobID,
	)
	return err
}

func markEpisodeJobDownloaded(jobID uint64, downloadEventID *uint64) error {
	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`UPDATE tv_episode_jobs
		SET status = 'downloaded',
			attempt_count = attempt_count + 1,
			last_checked_at = UTC_TIMESTAMP(),
			next_check_at = UTC_TIMESTAMP(),
			last_error = NULL,
			downloaded_download_event_id = COALESCE(?, downloaded_download_event_id),
			updated_at = UTC_TIMESTAMP()
		WHERE id = ?`,
		downloadEventID,
		jobID,
	)
	return err
}

func findLatestDownloadEventIDByFid(fid string, username string) (*uint64, error) {
	cleanFid := strings.TrimSpace(fid)
	cleanUsername := strings.TrimSpace(username)
	if cleanFid == "" || cleanUsername == "" {
		return nil, nil
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var eventID uint64
	err = db.QueryRowContext(
		ctx,
		`SELECT id
		FROM download_events
		WHERE success = 1
		  AND deleted_at IS NULL
		  AND fid = ?
		  AND username = ?
		ORDER BY created_at DESC
		LIMIT 1`,
		cleanFid,
		cleanUsername,
	).Scan(&eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &eventID, nil
}

func GetTvShowInstallStatus(username string, showID int64) (*TvShowInstallStatus, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	subscription, err := getTvShowSubscriptionByShowID(cleanUsername, showID)
	if err != nil {
		return nil, err
	}

	jobs, err := listTvEpisodeJobsByShow(cleanUsername, showID, 200)
	if err != nil {
		return nil, err
	}

	status := &TvShowInstallStatus{
		Subscription: subscription,
		Jobs:         jobs,
	}

	for _, job := range jobs {
		switch job.Status {
		case "downloaded":
			status.DownloadedCount++
		case "failed":
			status.FailedCount++
		default:
			status.PendingCount++
		}
	}

	return status, nil
}

func ConfigureTvShowAutoInstall(username string, show TvMazeShow, preferredQuality string, autoInstallUpcoming bool) (*TvShowSubscriptionRecord, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}
	if show.ID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	quality := NormalizeQualityPreference(preferredQuality)
	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return nil, err
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO tv_show_subscriptions (
			user_id,
			username,
			tvmaze_show_id,
			show_name,
			preferred_quality,
			auto_install_upcoming,
			enabled,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, 1, UTC_TIMESTAMP())
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			username = VALUES(username),
			show_name = VALUES(show_name),
			preferred_quality = VALUES(preferred_quality),
			auto_install_upcoming = VALUES(auto_install_upcoming),
			enabled = 1,
			updated_at = UTC_TIMESTAMP()`,
		userID,
		cleanUsername,
		show.ID,
		nullableString(show.Name),
		quality,
		autoInstallUpcoming,
	)
	if err != nil {
		return nil, err
	}

	subscriptionID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	prunedJobs, pruneErr := prunePendingSubscriptionJobsForOtherQualities(uint64(subscriptionID), quality)
	if pruneErr != nil {
		log.Printf("failed pruning pending jobs for subscription %d quality %s: %v", subscriptionID, quality, pruneErr)
	} else if prunedJobs > 0 {
		log.Printf("pruned %d pending jobs for subscription %d with non-%sp quality", prunedJobs, subscriptionID, quality)
	}

	if autoInstallUpcoming {
		if syncErr := queueUpcomingEpisodesForShow(cleanUsername, uint64(subscriptionID), show.ID, quality); syncErr != nil {
			log.Printf("failed to queue upcoming episodes for show %d: %v", show.ID, syncErr)
		}
		go func() {
			if processErr := ProcessDueTvEpisodeJobs(5); processErr != nil {
				log.Printf("failed to trigger tv episode worker after enabling auto install: %v", processErr)
			}
		}()
	}

	return getTvShowSubscriptionByShowID(cleanUsername, show.ID)
}

func QueueWholeShowInstall(username string, showID int64, preferredQuality string) (*TvInstallQueueResult, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	episodes, err := TvMazeGetEpisodes(showID)
	if err != nil {
		return nil, err
	}

	if len(episodes) == 0 {
		return &TvInstallQueueResult{}, nil
	}

	quality := NormalizeQualityPreference(preferredQuality)
	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return nil, err
	}

	show, err := TvMazeGetShow(showID)
	if err != nil {
		log.Printf("failed to load show %d for boxset matching, falling back to episode queue: %v", showID, err)
		return queueEpisodeBatchInstall(cleanUsername, showID, episodes, quality)
	}

	boxsetTorrents, err := TlSeriesBoxsetSearchByTvMaze(showID, showID)
	if err != nil {
		log.Printf("boxset search failed for show %d, falling back to episode queue: %v", showID, err)
		return queueEpisodeBatchInstall(cleanUsername, showID, episodes, quality)
	}

	episodesBySeason := make(map[int][]TvMazeEpisode)
	seasonNumbers := make([]int, 0)
	fallbackEpisodes := make([]TvMazeEpisode, 0)

	for _, episode := range episodes {
		if episode.Season <= 0 {
			fallbackEpisodes = append(fallbackEpisodes, episode)
			continue
		}

		if _, exists := episodesBySeason[episode.Season]; !exists {
			seasonNumbers = append(seasonNumbers, episode.Season)
		}
		episodesBySeason[episode.Season] = append(episodesBySeason[episode.Season], episode)
	}

	sort.Ints(seasonNumbers)

	result := &TvInstallQueueResult{}
	for _, seasonNumber := range seasonNumbers {
		seasonEpisodes := episodesBySeason[seasonNumber]
		matched, triggered, installErr := installSeasonBoxsetIfPossible(
			userID,
			cleanUsername,
			show.ID,
			show.Name,
			seasonNumber,
			quality,
			boxsetTorrents,
			seasonEpisodes,
		)
		if installErr != nil {
			log.Printf("boxset install failed for show %d season %d, falling back to episodes: %v", show.ID, seasonNumber, installErr)
			fallbackEpisodes = append(fallbackEpisodes, seasonEpisodes...)
			continue
		}
		if matched {
			if triggered {
				result.Queued += countInstallableEpisodes(seasonEpisodes)
				result.Triggered++
			}
			continue
		}

		fallbackEpisodes = append(fallbackEpisodes, seasonEpisodes...)
	}

	if len(fallbackEpisodes) > 0 {
		fallbackResult, queueErr := queueEpisodeBatchInstall(cleanUsername, showID, fallbackEpisodes, quality)
		if queueErr != nil {
			return nil, queueErr
		}

		result.Queued += fallbackResult.Queued
		result.Skipped += fallbackResult.Skipped
		result.Triggered += fallbackResult.Triggered
	}

	return result, nil
}

func QueueSeasonInstall(username string, showID int64, seasonNumber int, preferredQuality string) (*TvInstallQueueResult, error) {
	if seasonNumber <= 0 {
		return nil, fmt.Errorf("season number must be greater than zero")
	}

	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	episodes, err := TvMazeGetEpisodes(showID)
	if err != nil {
		return nil, err
	}

	seasonEpisodes := make([]TvMazeEpisode, 0)
	for _, episode := range episodes {
		if episode.Season != seasonNumber {
			continue
		}
		seasonEpisodes = append(seasonEpisodes, episode)
	}

	if len(seasonEpisodes) == 0 {
		return &TvInstallQueueResult{}, nil
	}

	quality := NormalizeQualityPreference(preferredQuality)
	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return nil, err
	}

	show, err := TvMazeGetShow(showID)
	if err != nil {
		log.Printf("failed to load show %d for season boxset matching, falling back to episode queue: %v", showID, err)
		return queueEpisodeBatchInstall(cleanUsername, showID, seasonEpisodes, quality)
	}

	boxsetTorrents, err := TlSeriesBoxsetSearchByTvMaze(showID, showID)
	if err != nil {
		log.Printf("boxset search failed for show %d season %d, falling back to episode queue: %v", showID, seasonNumber, err)
		return queueEpisodeBatchInstall(cleanUsername, showID, seasonEpisodes, quality)
	}

	matched, triggered, installErr := installSeasonBoxsetIfPossible(
		userID,
		cleanUsername,
		show.ID,
		show.Name,
		seasonNumber,
		quality,
		boxsetTorrents,
		seasonEpisodes,
	)
	if installErr != nil {
		log.Printf("boxset install failed for show %d season %d, falling back to episode queue: %v", showID, seasonNumber, installErr)
		return queueEpisodeBatchInstall(cleanUsername, showID, seasonEpisodes, quality)
	}
	if matched {
		result := &TvInstallQueueResult{
			Queued: 0,
		}
		if triggered {
			result.Queued = countInstallableEpisodes(seasonEpisodes)
			result.Triggered = 1
		}
		return result, nil
	}

	return queueEpisodeBatchInstall(cleanUsername, showID, seasonEpisodes, quality)
}

func QueueEpisodeInstall(username string, showID int64, episodeID int64, preferredQuality string) (*TvInstallQueueResult, error) {
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	episode, err := TvMazeGetEpisode(episodeID)
	if err != nil {
		return nil, err
	}

	result, err := queueEpisodeBatchInstall(username, showID, []TvMazeEpisode{*episode}, preferredQuality)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func queueEpisodeBatchInstall(username string, showID int64, episodes []TvMazeEpisode, preferredQuality string) (*TvInstallQueueResult, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}
	if showID <= 0 {
		return nil, fmt.Errorf("show id is required")
	}

	quality := NormalizeQualityPreference(preferredQuality)
	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return nil, err
	}

	result := &TvInstallQueueResult{}
	now := time.Now().UTC()
	for _, episode := range episodes {
		if !isInstallableEpisode(episode, now) {
			result.Skipped++
			continue
		}

		isDownloaded, lookupErr := isEpisodeJobDownloaded(userID, episode.ID, quality)
		if lookupErr != nil {
			log.Printf("failed checking existing status for episode %d quality %s: %v", episode.ID, quality, lookupErr)
			result.Skipped++
			continue
		}
		if isDownloaded {
			result.Skipped++
			continue
		}

		if _, queueErr := queueSingleEpisodeJob(userID, cleanUsername, nil, showID, episode, quality, true); queueErr != nil {
			log.Printf("failed queueing episode %d for show %d: %v", episode.ID, showID, queueErr)
			result.Skipped++
			continue
		}

		result.Queued++
	}

	if result.Queued > 0 {
		go func() {
			if processErr := ProcessDueTvEpisodeJobs(10); processErr != nil {
				log.Printf("failed to trigger tv episode worker after queueing jobs: %v", processErr)
			}
		}()
		result.Triggered = result.Queued
	}

	return result, nil
}

func queueUpcomingEpisodesForShow(username string, subscriptionID uint64, showID int64, quality string) error {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return fmt.Errorf("username is required")
	}
	if subscriptionID == 0 {
		return fmt.Errorf("subscription id is required")
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return err
	}

	episodes, err := TvMazeGetEpisodes(showID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, episode := range episodes {
		if strings.TrimSpace(episode.Airstamp) == "" {
			continue
		}

		airstamp, parseErr := ParseTvMazeAirstamp(episode.Airstamp)
		if parseErr != nil {
			continue
		}
		if !airstamp.After(now) {
			continue
		}

		subscriptionIDCopy := subscriptionID
		if _, queueErr := queueSingleEpisodeJob(userID, cleanUsername, &subscriptionIDCopy, showID, episode, quality, false); queueErr != nil {
			log.Printf("failed to queue upcoming episode %d for show %d: %v", episode.ID, showID, queueErr)
		}
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		`UPDATE tv_show_subscriptions
		SET last_synced_at = UTC_TIMESTAMP(),
			updated_at = UTC_TIMESTAMP()
		WHERE id = ?`,
		subscriptionID,
	)

	return err
}

func prunePendingSubscriptionJobsForOtherQualities(subscriptionID uint64, quality string) (int64, error) {
	if subscriptionID == 0 {
		return 0, fmt.Errorf("subscription id is required")
	}

	db, err := dbConn()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.ExecContext(
		ctx,
		`DELETE FROM tv_episode_jobs
		WHERE subscription_id = ?
		  AND status IN ('pending', 'searching')
		  AND preferred_quality <> ?`,
		subscriptionID,
		NormalizeQualityPreference(quality),
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func queueSingleEpisodeJob(userID uint64, username string, subscriptionID *uint64, showID int64, episode TvMazeEpisode, quality string, immediate bool) (uint64, error) {
	if episode.ID <= 0 {
		return 0, fmt.Errorf("episode id is required")
	}

	nextCheck := time.Now().UTC()
	var airstamp any
	if strings.TrimSpace(episode.Airstamp) != "" {
		if parsedAirstamp, err := ParseTvMazeAirstamp(episode.Airstamp); err == nil {
			airstamp = parsedAirstamp
			if !immediate {
				if calculatedNext, ok := NextEpisodeRetryTime(parsedAirstamp, time.Now().UTC()); ok {
					nextCheck = calculatedNext
				}
			}
		}
	}

	db, err := dbConn()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO tv_episode_jobs (
			subscription_id,
			user_id,
			username,
			tvmaze_show_id,
			tvmaze_episode_id,
			episode_name,
			season_number,
			episode_number,
			airstamp,
			preferred_quality,
			status,
			attempt_count,
			next_check_at,
			last_checked_at,
			last_error,
			downloaded_download_event_id,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 0, ?, NULL, NULL, NULL, UTC_TIMESTAMP())
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			subscription_id = COALESCE(VALUES(subscription_id), subscription_id),
			episode_name = VALUES(episode_name),
			season_number = VALUES(season_number),
			episode_number = VALUES(episode_number),
			airstamp = VALUES(airstamp),
			next_check_at = CASE
				WHEN status = 'downloaded'
					AND EXISTS (
						SELECT 1
						FROM download_events d
						WHERE d.id = downloaded_download_event_id
						  AND d.deleted_at IS NULL
					)
				THEN next_check_at
				ELSE LEAST(next_check_at, VALUES(next_check_at))
			END,
			status = CASE
				WHEN status = 'downloaded'
					AND EXISTS (
						SELECT 1
						FROM download_events d
						WHERE d.id = downloaded_download_event_id
						  AND d.deleted_at IS NULL
					)
				THEN status
				ELSE 'pending'
			END,
			downloaded_download_event_id = CASE
				WHEN status = 'downloaded'
					AND EXISTS (
						SELECT 1
						FROM download_events d
						WHERE d.id = downloaded_download_event_id
						  AND d.deleted_at IS NULL
					)
				THEN downloaded_download_event_id
				ELSE NULL
			END,
			last_error = NULL,
			updated_at = UTC_TIMESTAMP()`,
		nullableUint64Ptr(subscriptionID),
		userID,
		username,
		showID,
		episode.ID,
		nullableString(episode.Name),
		nullableInt(episode.Season),
		nullableInt(episode.Number),
		airstamp,
		NormalizeQualityPreference(quality),
		nextCheck,
	)
	if err != nil {
		return 0, err
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(jobID), nil
}

func getTvShowSubscriptionByShowID(username string, showID int64) (*TvShowSubscriptionRecord, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var record TvShowSubscriptionRecord
	var updatedAt time.Time
	err = db.QueryRowContext(
		ctx,
		`SELECT
			id,
			tvmaze_show_id,
			COALESCE(show_name, ''),
			preferred_quality,
			auto_install_upcoming,
			enabled,
			updated_at
		FROM tv_show_subscriptions
		WHERE username = ?
		  AND tvmaze_show_id = ?
		LIMIT 1`,
		cleanUsername,
		showID,
	).Scan(
		&record.ID,
		&record.TvMazeShowID,
		&record.ShowName,
		&record.PreferredQuality,
		&record.AutoInstallUpcoming,
		&record.Enabled,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	record.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	record.PreferredQuality = NormalizeQualityPreference(record.PreferredQuality)

	return &record, nil
}

func listTvEpisodeJobsByShow(username string, showID int64, limit int) ([]TvEpisodeJobRecord, error) {
	if limit <= 0 {
		limit = 200
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT
			j.id,
			j.tvmaze_show_id,
			j.tvmaze_episode_id,
			COALESCE(j.episode_name, ''),
			COALESCE(j.season_number, 0),
			COALESCE(j.episode_number, 0),
			j.airstamp,
			j.preferred_quality,
			j.status,
			j.attempt_count,
			j.next_check_at,
			j.last_checked_at,
			COALESCE(j.last_error, '')
		FROM tv_episode_jobs j
		LEFT JOIN download_events d
			ON d.id = j.downloaded_download_event_id
		WHERE j.username = ?
		  AND j.tvmaze_show_id = ?
		  AND NOT (
			j.status = 'downloaded'
			AND j.downloaded_download_event_id IS NOT NULL
			AND d.deleted_at IS NOT NULL
		  )
		ORDER BY j.season_number ASC, j.episode_number ASC, j.tvmaze_episode_id ASC
		LIMIT ?`,
		username,
		showID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TvEpisodeJobRecord, 0)
	for rows.Next() {
		var (
			record        TvEpisodeJobRecord
			airstamp      sql.NullTime
			nextCheckAt   time.Time
			lastCheckedAt sql.NullTime
		)

		if err := rows.Scan(
			&record.ID,
			&record.TvMazeShowID,
			&record.TvMazeEpisodeID,
			&record.EpisodeName,
			&record.SeasonNumber,
			&record.EpisodeNumber,
			&airstamp,
			&record.PreferredQuality,
			&record.Status,
			&record.AttemptCount,
			&nextCheckAt,
			&lastCheckedAt,
			&record.LastError,
		); err != nil {
			return nil, err
		}

		nextCheck := nextCheckAt.UTC().Format(time.RFC3339)
		record.NextCheckAt = nextCheck
		record.PreferredQuality = NormalizeQualityPreference(record.PreferredQuality)
		if airstamp.Valid {
			value := airstamp.Time.UTC().Format(time.RFC3339)
			record.Airstamp = &value
		}
		if lastCheckedAt.Valid {
			value := lastCheckedAt.Time.UTC().Format(time.RFC3339)
			record.LastCheckedAt = &value
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableUint64Ptr(value *uint64) any {
	if value == nil || *value == 0 {
		return nil
	}
	return *value
}

func BuildTvMazeDownloadDataFromTorrent(torrent TlSeriesTorrent, showID int64, episodeID int64) models.DownloadData {
	data := ConvertTlSeriesTorrentToDownloadData(torrent, showID, episodeID)
	if !IsMovieOrTVCategory(data.CategoryID) {
		data.CategoryID = tvDefaultCategoryID
	}
	return data
}

func installSeasonBoxsetIfPossible(
	userID uint64,
	username string,
	showID int64,
	showName string,
	seasonNumber int,
	quality string,
	boxsetTorrents []TlSeriesTorrent,
	seasonEpisodes []TvMazeEpisode,
) (bool, bool, error) {
	selected := SelectBestBoxsetTorrentByQuality(
		boxsetTorrents,
		showName,
		seasonNumber,
		quality,
	)
	if selected == nil {
		return false, false, nil
	}

	downloadData := BuildTvMazeDownloadDataFromTorrent(*selected, showID, 0)
	downloadData.TvMazeEpisodeID = ""

	isDownloaded, err := IsFidAlreadyDownloaded(downloadData.Fid)
	if err != nil {
		return true, false, err
	}
	downloadEventID, eventLookupErr := findLatestDownloadEventIDByFid(downloadData.Fid, username)
	if eventLookupErr != nil {
		log.Printf("failed finding existing boxset download event for fid %s: %v", downloadData.Fid, eventLookupErr)
	}

	if isDownloaded {
		if markErr := markEpisodesDownloadedFromBoxset(userID, username, showID, seasonEpisodes, quality, downloadEventID); markErr != nil {
			return true, false, markErr
		}
		return true, false, nil
	}

	qbtHash, err := TlDownloadRequest(downloadData)
	if err != nil {
		return true, false, err
	}

	if logErr := LogDownloadEvent(username, downloadData, qbtHash, true, "", "", "tv-boxset-install"); logErr != nil {
		log.Printf("failed to write boxset download event for show %d season %d: %v", showID, seasonNumber, logErr)
	}
	downloadEventID, eventLookupErr = findLatestDownloadEventIDByFid(downloadData.Fid, username)
	if eventLookupErr != nil {
		log.Printf("failed finding boxset download event for fid %s: %v", downloadData.Fid, eventLookupErr)
	}

	if markErr := markEpisodesDownloadedFromBoxset(userID, username, showID, seasonEpisodes, quality, downloadEventID); markErr != nil {
		return true, false, markErr
	}
	ScheduleAutoPlexScanForDownload(qbtHash, downloadData.CategoryID)

	return true, true, nil
}

func countInstallableEpisodes(episodes []TvMazeEpisode) int {
	now := time.Now().UTC()
	count := 0
	for _, episode := range episodes {
		if isInstallableEpisode(episode, now) {
			count++
		}
	}
	return count
}

func markEpisodesDownloadedFromBoxset(
	userID uint64,
	username string,
	showID int64,
	episodes []TvMazeEpisode,
	quality string,
	downloadEventID *uint64,
) error {
	if userID == 0 {
		return fmt.Errorf("user id is required")
	}

	now := time.Now().UTC()
	for _, episode := range episodes {
		if !isInstallableEpisode(episode, now) {
			continue
		}

		alreadyDownloaded, err := isEpisodeJobDownloaded(userID, episode.ID, quality)
		if err != nil {
			return err
		}
		if alreadyDownloaded {
			continue
		}

		jobID, err := queueSingleEpisodeJob(userID, username, nil, showID, episode, quality, true)
		if err != nil {
			return err
		}
		if err := markEpisodeJobDownloaded(jobID, downloadEventID); err != nil {
			return err
		}
	}

	return nil
}

func isEpisodeJobDownloaded(userID uint64, episodeID int64, quality string) (bool, error) {
	if userID == 0 {
		return false, fmt.Errorf("user id is required")
	}
	if episodeID <= 0 {
		return false, nil
	}

	db, err := dbConn()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err = db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM tv_episode_jobs j
		LEFT JOIN download_events d
			ON d.id = j.downloaded_download_event_id
		WHERE j.user_id = ?
		  AND j.tvmaze_episode_id = ?
		  AND j.preferred_quality = ?
		  AND j.status = 'downloaded'
		  AND (
			j.downloaded_download_event_id IS NULL
			OR d.deleted_at IS NULL
		  )`,
		userID,
		episodeID,
		NormalizeQualityPreference(quality),
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func isInstallableEpisode(episode TvMazeEpisode, now time.Time) bool {
	if episode.ID <= 0 {
		return false
	}

	if parsedAirstamp, err := ParseTvMazeAirstamp(strings.TrimSpace(episode.Airstamp)); err == nil {
		return !parsedAirstamp.After(now)
	}

	airdate := strings.TrimSpace(episode.Airdate)
	if airdate == "" {
		return false
	}

	parsedAirdate, err := time.Parse("2006-01-02", airdate)
	if err != nil {
		return false
	}

	return !parsedAirdate.After(now.UTC())
}
