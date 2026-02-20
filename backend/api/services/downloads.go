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
)

func ListDownloadEvents(username string, isAdmin bool) ([]models.DownloadEventRecord, error) {
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

	rows, err := db.QueryContext(
		ctx,
		`SELECT
			d.id,
			d.user_id,
			COALESCE(d.username, ''),
			COALESCE(d.fid, ''),
			COALESCE(d.filename, ''),
			COALESCE(d.category_id, 0),
			COALESCE(d.torrent_size, 0),
			d.is_freeleech,
			d.created_at,
			d.deleted_at,
			d.deleted_by_username,
			EXISTS(
				SELECT 1
				FROM download_delete_requests r
				WHERE r.download_event_id = d.id
				  AND r.status = 'pending'
			) AS has_pending_delete_request
		FROM download_events d
		WHERE d.success = 1
		  AND (? OR d.username = ?)
		ORDER BY d.created_at DESC`,
		isAdmin,
		cleanUsername,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	downloads := make([]models.DownloadEventRecord, 0)
	for rows.Next() {
		var (
			record            models.DownloadEventRecord
			userID            sql.NullInt64
			categoryID        sql.NullInt64
			torrentSize       sql.NullInt64
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
			&categoryID,
			&torrentSize,
			&record.IsFreeleech,
			&createdAt,
			&deletedAt,
			&deletedByUsername,
			&record.HasPendingDelete,
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

	return downloads, nil
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

	if isAdmin || isFreeleech {
		if strings.TrimSpace(qbtHash.String) == "" {
			return "", ErrDownloadMissingHash
		}

		if err := QbtDelete(qbtHash.String); err != nil {
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
			  AND status = 'pending'`,
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
				approved_at
			) VALUES (?, ?, ?, 'approved', ?, ?, ?, NOW())`,
				downloadID,
				userID,
				cleanUsername,
				nullableString(reason),
				userID,
				cleanUsername,
			); err != nil {
				return "", err
			}
		}

		if err := tx.Commit(); err != nil {
			return "", err
		}

		return DownloadDeleteActionDeleted, nil
	}

	var hasPending bool
	if err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM download_delete_requests
			WHERE download_event_id = ?
			  AND status = 'pending'
		)`,
		downloadID,
	).Scan(&hasPending); err != nil {
		return "", err
	}

	if hasPending {
		return "", ErrDeleteRequestAlreadyPending
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO download_delete_requests (
			download_event_id,
			requested_by_user_id,
			requested_by_username,
			status,
			request_note
		) VALUES (?, ?, ?, 'pending', ?)`,
		downloadID,
		userID,
		cleanUsername,
		nullableString(reason),
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return DownloadDeleteActionRequested, nil
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
			id,
			download_event_id,
			requested_by_username,
			status,
			COALESCE(request_note, ''),
			COALESCE(approved_by_username, ''),
			created_at,
			approved_at
		FROM download_delete_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]models.DownloadDeleteRequestRecord, 0)
	for rows.Next() {
		var (
			record     models.DownloadDeleteRequestRecord
			createdAt  time.Time
			approvedAt sql.NullTime
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
		); err != nil {
			return nil, err
		}

		record.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.UTC().Format(time.RFC3339)
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

	if status != "pending" {
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
