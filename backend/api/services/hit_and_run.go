package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	hitAndRunWorkerInterval = 10 * time.Minute
	hitAndRunSystemActor    = "system-hit-and-run"
)

var hitAndRunWorkerOnce sync.Once

func StartHitAndRunWorker() {
	hitAndRunWorkerOnce.Do(func() {
		go runHitAndRunWorker()
	})
}

func runHitAndRunWorker() {
	ticker := time.NewTicker(hitAndRunWorkerInterval)
	defer ticker.Stop()

	if err := ProcessDueHitAndRunDeletes(); err != nil {
		log.Printf("hit-and-run worker initial run failed: %v", err)
	}

	for range ticker.C {
		if err := ProcessDueHitAndRunDeletes(); err != nil {
			log.Printf("hit-and-run worker run failed: %v", err)
		}
	}
}

type hitAndRunDueRow struct {
	requestID       uint64
	downloadEventID uint64
	qbtHash         string
}

func ProcessDueHitAndRunDeletes() error {
	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT r.id, r.download_event_id, COALESCE(e.qbt_hash, '')
		FROM download_delete_requests r
		INNER JOIN download_events e ON e.id = r.download_event_id
		WHERE r.status = 'hit_and_run'
		  AND r.auto_delete_at IS NOT NULL
		  AND r.auto_delete_at <= UTC_TIMESTAMP()
		  AND e.deleted_at IS NULL
		ORDER BY r.auto_delete_at ASC
		LIMIT 100`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	due := make([]hitAndRunDueRow, 0)
	for rows.Next() {
		var row hitAndRunDueRow
		if err := rows.Scan(&row.requestID, &row.downloadEventID, &row.qbtHash); err != nil {
			return err
		}
		due = append(due, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, row := range due {
		if err := executeHitAndRunAutoDelete(row); err != nil {
			log.Printf("hit-and-run auto-delete failed for request %d: %v", row.requestID, err)
		}
	}

	return nil
}

func executeHitAndRunAutoDelete(row hitAndRunDueRow) error {
	cleanHash := strings.TrimSpace(row.qbtHash)
	if cleanHash == "" {
		return errors.New("download is missing qbt hash")
	}

	if err := QbtDelete(cleanHash); err != nil {
		return err
	}

	db, err := dbConn()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var deletedAt sql.NullTime
	if err := tx.QueryRowContext(
		ctx,
		`SELECT deleted_at FROM download_events WHERE id = ? FOR UPDATE`,
		row.downloadEventID,
	).Scan(&deletedAt); err != nil {
		return err
	}
	if deletedAt.Valid {
		return tx.Commit()
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE download_events
		SET deleted_at = NOW(),
			deleted_by_user_id = NULL,
			deleted_by_username = ?
		WHERE id = ?`,
		hitAndRunSystemActor,
		row.downloadEventID,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE download_delete_requests
		SET status = 'approved',
			approved_by_user_id = NULL,
			approved_by_username = ?,
			approved_at = NOW(),
			updated_at = NOW()
		WHERE id = ?`,
		hitAndRunSystemActor,
		row.requestID,
	); err != nil {
		return err
	}

	return tx.Commit()
}
