package services

import (
	"api/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDownloadNotFound            = errors.New("download not found")
	ErrDeleteNotAllowed            = errors.New("delete is not allowed for this download")
	ErrDownloadAlreadyDeleted      = errors.New("download is already deleted")
	ErrDeleteRequestAlreadyPending = errors.New("a delete request is already pending for this download")
	ErrDeleteRequestNotFound       = errors.New("delete request not found")
	ErrDeleteRequestNotPending     = errors.New("delete request is not pending")
	ErrDownloadMissingHash         = errors.New("download is missing qbt hash and cannot be deleted")
	ErrAdminRequired               = errors.New("admin role is required")
)

const (
	DownloadDeleteActionDeleted   = "deleted"
	DownloadDeleteActionRequested = "delete_requested"

	seedRequiredDuration    = 168 * time.Hour
	autoDeleteGraceDuration = 24 * time.Hour
)

type hitAndRunWindow struct {
	completionAt time.Time
	safeAt       time.Time
	autoDeleteAt time.Time
	hasWindow    bool
}

func computeHitAndRunWindow(completionUnix int) hitAndRunWindow {
	if completionUnix <= 0 {
		return hitAndRunWindow{}
	}
	completionAt := time.Unix(int64(completionUnix), 0).UTC()
	safeAt := completionAt.Add(seedRequiredDuration)
	return hitAndRunWindow{
		completionAt: completionAt,
		safeAt:       safeAt,
		autoDeleteAt: safeAt.Add(autoDeleteGraceDuration),
		hasWindow:    true,
	}
}

func IsFidAlreadyDownloaded(fid string) (bool, error) {
	cleanFid := strings.TrimSpace(fid)
	if cleanFid == "" {
		return false, fmt.Errorf("fid is required")
	}

	db, err := dbConn()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM download_events
			WHERE fid = ?
			  AND success = 1
			  AND deleted_at IS NULL
		)`,
		cleanFid,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func MarkSearchResultsWithDownloaded(resp map[string]any) error {
	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT fid
		FROM download_events
		WHERE fid IS NOT NULL
		  AND fid <> ''
		  AND success = 1
		  AND deleted_at IS NULL`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	downloadedFids := make(map[string]struct{})
	for rows.Next() {
		var fid string
		if err := rows.Scan(&fid); err != nil {
			return err
		}

		cleanFid := strings.TrimSpace(fid)
		if cleanFid == "" {
			continue
		}
		downloadedFids[cleanFid] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	torrents, ok := resp["torrentList"].([]any)
	if !ok {
		return nil
	}

	for _, raw := range torrents {
		torrent, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		fidValue, hasFid := torrent["fid"]
		if !hasFid {
			torrent["isDownloaded"] = false
			continue
		}

		fid := strings.TrimSpace(fmt.Sprintf("%v", fidValue))
		_, isDownloaded := downloadedFids[fid]
		torrent["isDownloaded"] = isDownloaded
	}

	return nil
}

type DownloadListParams struct {
	Status string
	Query  string
	User   string
	Sort   string
	Dir    string
	Limit  int
	Offset int
}

type DownloadListResult struct {
	Downloads      []models.DownloadEventRecord
	Total          int64
	AvailableUsers []string
}

func downloadOrderBy(sort, dir string) string {
	column := "d.created_at"
	switch sort {
	case "filename":
		column = "COALESCE(NULLIF(d.filename, ''), d.fid)"
	case "torrentSize":
		column = "COALESCE(d.torrent_size, 0)"
	case "username":
		column = "COALESCE(d.username, '')"
	case "deletedAt":
		column = "d.deleted_at"
	}
	direction := "DESC"
	if strings.EqualFold(dir, "asc") {
		direction = "ASC"
	}
	return fmt.Sprintf("ORDER BY %s %s, d.id %s", column, direction, direction)
}

func ListDownloadEvents(username string, isAdmin bool, params DownloadListParams) (*DownloadListResult, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	limit := params.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	where := []string{"d.success = 1"}
	args := []any{}

	if strings.EqualFold(strings.TrimSpace(params.Status), "deleted") {
		where = append(where, "d.deleted_at IS NOT NULL")
	} else {
		where = append(where, "d.deleted_at IS NULL")
	}

	if isAdmin {
		if cleanUserFilter := strings.TrimSpace(params.User); cleanUserFilter != "" {
			where = append(where, "LOWER(d.username) = ?")
			args = append(args, strings.ToLower(cleanUserFilter))
		}
	} else {
		where = append(where, "d.username = ?")
		args = append(args, cleanUsername)
	}

	if cleanQuery := strings.TrimSpace(params.Query); cleanQuery != "" {
		pattern := "%" + cleanQuery + "%"
		where = append(where, "(d.filename LIKE ? OR d.fid LIKE ?)")
		args = append(args, pattern, pattern)
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var total int64
	if err := db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM download_events d "+whereSQL,
		args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	availableUsers := make([]string, 0)
	if isAdmin {
		userRows, err := db.QueryContext(
			ctx,
			`SELECT DISTINCT username
			FROM download_events
			WHERE success = 1
			  AND username IS NOT NULL
			  AND username <> ''
			ORDER BY username ASC`,
		)
		if err != nil {
			return nil, err
		}
		for userRows.Next() {
			var name string
			if err := userRows.Scan(&name); err != nil {
				userRows.Close()
				return nil, err
			}
			availableUsers = append(availableUsers, name)
		}
		userRows.Close()
	}

	listQuery := `SELECT
			d.id,
			d.user_id,
			COALESCE(d.username, ''),
			COALESCE(d.fid, ''),
			COALESCE(d.filename, ''),
			COALESCE(d.tvmaze_id, ''),
			COALESCE(d.tvmaze_episode_id, ''),
			COALESCE(d.category_id, 0),
			COALESCE(d.torrent_size, 0),
			d.is_freeleech,
			COALESCE(d.qbt_hash, ''),
			d.created_at,
			d.deleted_at,
			d.deleted_by_username,
			EXISTS(
				SELECT 1
				FROM download_delete_requests r
				WHERE r.download_event_id = d.id
				  AND r.status = 'pending'
			) AS has_pending_delete_request,
			EXISTS(
				SELECT 1
				FROM download_delete_requests r
				WHERE r.download_event_id = d.id
				  AND r.status = 'hit_and_run'
			) AS has_hit_and_run
		FROM download_events d
		` + whereSQL + ` ` + downloadOrderBy(params.Sort, params.Dir) + ` LIMIT ? OFFSET ?`

	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	downloads := make([]models.DownloadEventRecord, 0, limit)
	for rows.Next() {
		var (
			record            models.DownloadEventRecord
			userID            sql.NullInt64
			categoryID        sql.NullInt64
			torrentSize       sql.NullInt64
			qbtHash           string
			createdAt         time.Time
			deletedAt         sql.NullTime
			deletedByUsername sql.NullString
		)

		if err := rows.Scan(
			&record.ID,
			&userID,
			&record.Username,
			&record.Fid,
			&record.Filename,
			&record.TvMazeID,
			&record.TvMazeEpisodeID,
			&categoryID,
			&torrentSize,
			&record.IsFreeleech,
			&qbtHash,
			&createdAt,
			&deletedAt,
			&deletedByUsername,
			&record.HasPendingDelete,
			&record.HasHitAndRun,
		); err != nil {
			return nil, err
		}

		if userID.Valid && userID.Int64 > 0 {
			value := uint64(userID.Int64)
			record.UserID = &value
		}

		if categoryID.Valid {
			record.CategoryID = int(categoryID.Int64)
		}

		if torrentSize.Valid && torrentSize.Int64 > 0 {
			record.TorrentSize = uint64(torrentSize.Int64)
		}

		record.QbtHash = strings.TrimSpace(qbtHash)
		record.CreatedAt = createdAt.UTC().Format(time.RFC3339)

		if deletedAt.Valid {
			formatted := deletedAt.Time.UTC().Format(time.RFC3339)
			record.DeletedAt = &formatted
		}

		if deletedByUsername.Valid {
			value := deletedByUsername.String
			record.DeletedByUsername = &value
		}

		downloads = append(downloads, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(downloads) == 0 {
		return &DownloadListResult{Downloads: downloads, Total: total, AvailableUsers: availableUsers}, nil
	}

	// Skip the live qBit enrichment when listing deleted rows — their qBit
	// torrents are gone, the state column already reflects that.
	if strings.EqualFold(strings.TrimSpace(params.Status), "deleted") {
		for i := range downloads {
			downloads[i].ProgressPercent = 100
			downloads[i].QbtState = "deleted"
		}
		return &DownloadListResult{Downloads: downloads, Total: total, AvailableUsers: availableUsers}, nil
	}

	torrentsByHash, err := QbtGetAllTorrentsByHash()
	if err == nil {
		for i := range downloads {
			hash := strings.ToLower(strings.TrimSpace(downloads[i].QbtHash))
			if hash == "" {
				continue
			}

			if torrent, exists := torrentsByHash[hash]; exists {
				downloads[i].QbtState = strings.TrimSpace(torrent.State)
				downloads[i].ProgressPercent = torrent.Progress * 100
				if downloads[i].TorrentSize == 0 && torrent.Size > 0 {
					downloads[i].TorrentSize = uint64(torrent.Size)
				}

				window := computeHitAndRunWindow(torrent.Completion_on)
				if window.hasWindow {
					completedAt := window.completionAt.Format(time.RFC3339)
					safeAt := window.safeAt.Format(time.RFC3339)
					downloads[i].CompletedAt = &completedAt
					downloads[i].SafeToDeleteAt = &safeAt
				}
				continue
			}

			if downloads[i].DeletedAt != nil {
				downloads[i].ProgressPercent = 100
				downloads[i].QbtState = "deleted"
			} else {
				downloads[i].QbtState = "missing"
			}
		}
	}

	return &DownloadListResult{Downloads: downloads, Total: total, AvailableUsers: availableUsers}, nil
}

func isActiveTorrentState(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "downloading", "stalleddl", "forceddl", "metadl", "queueddl", "checkingdl", "allocating":
		return true
	default:
		return false
	}
}

func DeleteOrRequestDownload(downloadID uint64, username string, isAdmin bool, reason string) (string, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return "", fmt.Errorf("username is required")
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return "", err
	}

	db, err := dbConn()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var (
		ownerUsername string
		isFreeleech   bool
		qbtHash       sql.NullString
		deletedAt     sql.NullTime
	)

	err = tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(username, ''), is_freeleech, qbt_hash, deleted_at
		FROM download_events
		WHERE id = ? AND success = 1
		FOR UPDATE`,
		downloadID,
	).Scan(
		&ownerUsername,
		&isFreeleech,
		&qbtHash,
		&deletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDownloadNotFound
	}
	if err != nil {
		return "", err
	}

	if deletedAt.Valid {
		return "", ErrDownloadAlreadyDeleted
	}

	if !isAdmin && !strings.EqualFold(strings.TrimSpace(ownerUsername), cleanUsername) {
		return "", ErrDeleteNotAllowed
	}

	cleanHash := strings.TrimSpace(qbtHash.String)
	if cleanHash == "" {
		return "", ErrDownloadMissingHash
	}

	torrent, lookupErr := QbtGetTorrentByHash(cleanHash)
	if lookupErr != nil {
		return "", lookupErr
	}

	var window hitAndRunWindow
	if torrent != nil {
		window = computeHitAndRunWindow(torrent.Completion_on)
	}

	now := time.Now().UTC()
	pastAutoDeleteWindow := window.hasWindow && !now.Before(window.autoDeleteAt)

	// Instant-delete eligibility:
	//   - admin (always), OR
	//   - regular user on a freeleech torrent that's past the 168+24h window.
	canInstantDelete := isAdmin || (isFreeleech && pastAutoDeleteWindow)

	if canInstantDelete {
		if err := QbtDelete(cleanHash); err != nil {
			return "", err
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE download_events
			SET deleted_at = NOW(),
				deleted_by_user_id = ?,
				deleted_by_username = ?
			WHERE id = ?`,
			userID,
			cleanUsername,
			downloadID,
		); err != nil {
			return "", err
		}

		updatedRequestsResult, err := tx.ExecContext(
			ctx,
			`UPDATE download_delete_requests
			SET status = 'approved',
				approved_by_user_id = ?,
				approved_by_username = ?,
				approved_at = NOW(),
				updated_at = NOW()
			WHERE download_event_id = ?
			  AND status IN ('pending', 'hit_and_run')`,
			userID,
			cleanUsername,
			downloadID,
		)
		if err != nil {
			return "", err
		}

		updatedRequestsCount, err := updatedRequestsResult.RowsAffected()
		if err != nil {
			return "", err
		}

		if updatedRequestsCount == 0 {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO download_delete_requests (
				download_event_id,
				requested_by_user_id,
				requested_by_username,
				status,
				request_note,
				approved_by_user_id,
				approved_by_username,
				approved_at,
				safe_to_delete_at,
				auto_delete_at
			) VALUES (?, ?, ?, 'approved', ?, ?, ?, NOW(), ?, ?)`,
				downloadID,
				userID,
				cleanUsername,
				nullableString(reason),
				userID,
				cleanUsername,
				nullableHitAndRunTime(window, "safe"),
				nullableHitAndRunTime(window, "auto"),
			); err != nil {
				return "", err
			}
		}

		if err := tx.Commit(); err != nil {
			return "", err
		}

		return DownloadDeleteActionDeleted, nil
	}

	// Non-admin path: if still downloading, pause it so we stop accruing
	// download stats while waiting on admin approval.
	if torrent != nil && isActiveTorrentState(torrent.State) {
		if pauseErr := QbtPause(cleanHash); pauseErr != nil {
			fmt.Printf("failed to pause torrent %s during delete request: %v\n", cleanHash, pauseErr)
		}
	}

	var hasOpen bool
	if err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM download_delete_requests
			WHERE download_event_id = ?
			  AND status IN ('pending', 'hit_and_run')
		)`,
		downloadID,
	).Scan(&hasOpen); err != nil {
		return "", err
	}

	if hasOpen {
		// Treat duplicate clicks as a no-op success — the request is already
		// queued. The caller's intent (mark for delete) is satisfied.
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return DownloadDeleteActionRequested, nil
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO download_delete_requests (
			download_event_id,
			requested_by_user_id,
			requested_by_username,
			status,
			request_note,
			safe_to_delete_at,
			auto_delete_at
		) VALUES (?, ?, ?, 'pending', ?, ?, ?)`,
		downloadID,
		userID,
		cleanUsername,
		nullableString(reason),
		nullableHitAndRunTime(window, "safe"),
		nullableHitAndRunTime(window, "auto"),
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return DownloadDeleteActionRequested, nil
}

func nullableHitAndRunTime(window hitAndRunWindow, kind string) any {
	if !window.hasWindow {
		return nil
	}
	switch kind {
	case "safe":
		return window.safeAt
	case "auto":
		return window.autoDeleteAt
	default:
		return nil
	}
}

func ListResolvedDeleteRequests(username string, isAdmin bool, limit int) ([]models.DownloadDeleteRequestRecord, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	if limit <= 0 || limit > 500 {
		limit = 100
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT
			r.id,
			r.download_event_id,
			r.requested_by_username,
			r.status,
			COALESCE(r.request_note, ''),
			COALESCE(r.approved_by_username, ''),
			r.created_at,
			r.approved_at,
			r.safe_to_delete_at,
			r.auto_delete_at,
			COALESCE(e.filename, ''),
			COALESCE(e.fid, ''),
			COALESCE(e.torrent_size, 0),
			COALESCE(e.is_freeleech, 0)
		FROM download_delete_requests r
		LEFT JOIN download_events e ON e.id = r.download_event_id
		WHERE r.status IN ('approved', 'rejected')`

	var args []any
	if !isAdmin {
		query += ` AND r.requested_by_username = ?`
		args = append(args, cleanUsername)
	}
	query += ` ORDER BY COALESCE(r.approved_at, r.created_at) DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeleteRequestRows(rows)
}

func ListPendingDeleteRequests(username string, isAdmin bool) ([]models.DownloadDeleteRequestRecord, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	if !isAdmin {
		return nil, ErrAdminRequired
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
			r.id,
			r.download_event_id,
			r.requested_by_username,
			r.status,
			COALESCE(r.request_note, ''),
			COALESCE(r.approved_by_username, ''),
			r.created_at,
			r.approved_at,
			r.safe_to_delete_at,
			r.auto_delete_at,
			COALESCE(e.filename, ''),
			COALESCE(e.fid, ''),
			COALESCE(e.torrent_size, 0),
			COALESCE(e.is_freeleech, 0)
		FROM download_delete_requests r
		LEFT JOIN download_events e ON e.id = r.download_event_id
		WHERE r.status = 'pending'
		ORDER BY r.created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeleteRequestRows(rows)
}

func ListHitAndRunRequests(username string, isAdmin bool) ([]models.DownloadDeleteRequestRecord, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return nil, fmt.Errorf("username is required")
	}

	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT
			r.id,
			r.download_event_id,
			r.requested_by_username,
			r.status,
			COALESCE(r.request_note, ''),
			COALESCE(r.approved_by_username, ''),
			r.created_at,
			r.approved_at,
			r.safe_to_delete_at,
			r.auto_delete_at,
			COALESCE(e.filename, ''),
			COALESCE(e.fid, ''),
			COALESCE(e.torrent_size, 0),
			COALESCE(e.is_freeleech, 0)
		FROM download_delete_requests r
		LEFT JOIN download_events e ON e.id = r.download_event_id
		WHERE r.status = 'hit_and_run'`

	var args []any
	if !isAdmin {
		query += ` AND r.requested_by_username = ?`
		args = append(args, cleanUsername)
	}
	query += ` ORDER BY COALESCE(r.safe_to_delete_at, r.created_at) ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDeleteRequestRows(rows)
}

func scanDeleteRequestRows(rows *sql.Rows) ([]models.DownloadDeleteRequestRecord, error) {
	requests := make([]models.DownloadDeleteRequestRecord, 0)
	for rows.Next() {
		var (
			record         models.DownloadDeleteRequestRecord
			createdAt      time.Time
			approvedAt     sql.NullTime
			safeToDeleteAt sql.NullTime
			autoDeleteAt   sql.NullTime
		)

		if err := rows.Scan(
			&record.ID,
			&record.DownloadEventID,
			&record.RequestedByUsername,
			&record.Status,
			&record.Reason,
			&record.ApprovedByUsername,
			&createdAt,
			&approvedAt,
			&safeToDeleteAt,
			&autoDeleteAt,
			&record.DownloadFilename,
			&record.DownloadFid,
			&record.DownloadSize,
			&record.DownloadIsFreeleech,
		); err != nil {
			return nil, err
		}

		record.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.UTC().Format(time.RFC3339)
		}
		if safeToDeleteAt.Valid {
			value := safeToDeleteAt.Time.UTC().Format(time.RFC3339)
			record.SafeToDeleteAt = &value
		}
		if autoDeleteAt.Valid {
			value := autoDeleteAt.Time.UTC().Format(time.RFC3339)
			record.AutoDeleteAt = &value
		}

		requests = append(requests, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func ApproveDeleteRequest(requestID uint64, username string, isAdmin bool) error {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return fmt.Errorf("username is required")
	}

	if !isAdmin {
		return ErrAdminRequired
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return err
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		downloadEventID uint64
		status          string
		qbtHash         sql.NullString
		deletedAt       sql.NullTime
	)

	err = tx.QueryRowContext(
		ctx,
		`SELECT
			r.download_event_id,
			r.status,
			d.qbt_hash,
			d.deleted_at
		FROM download_delete_requests r
		INNER JOIN download_events d ON d.id = r.download_event_id
		WHERE r.id = ?
		FOR UPDATE`,
		requestID,
	).Scan(
		&downloadEventID,
		&status,
		&qbtHash,
		&deletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDeleteRequestNotFound
	}
	if err != nil {
		return err
	}

	if status != "pending" && status != "hit_and_run" {
		return ErrDeleteRequestNotPending
	}

	if deletedAt.Valid {
		return ErrDownloadAlreadyDeleted
	}

	if strings.TrimSpace(qbtHash.String) == "" {
		return ErrDownloadMissingHash
	}

	if err := QbtDelete(qbtHash.String); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE download_events
		SET deleted_at = NOW(),
			deleted_by_user_id = ?,
			deleted_by_username = ?
		WHERE id = ?`,
		userID,
		cleanUsername,
		downloadEventID,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE download_delete_requests
		SET status = 'approved',
			approved_by_user_id = ?,
			approved_by_username = ?,
			approved_at = NOW(),
			updated_at = NOW()
		WHERE id = ?`,
		userID,
		cleanUsername,
		requestID,
	); err != nil {
		return err
	}

	return tx.Commit()
}
