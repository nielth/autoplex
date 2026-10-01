package services

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const qbtImportUserAgent = "qbt-import"

// ImportQbtTorrents adds every torrent in qBittorrent that autoplex does not
// track yet as a download by username, dated with qbt's added time. Torrents
// already tracked (by hash, deleted or not) are skipped, so it is safe to rerun.
func ImportQbtTorrents(username string) (int, error) {
	cleanUsername := strings.TrimSpace(username)
	if cleanUsername == "" {
		return 0, fmt.Errorf("username is required")
	}

	torrentsByHash, err := QbtGetAllTorrentsByHash()
	if err != nil {
		return 0, err
	}

	userID, err := ensureUserByUsername(cleanUsername)
	if err != nil {
		return 0, err
	}

	db, err := dbConn()
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `SELECT LOWER(qbt_hash) FROM download_events WHERE qbt_hash IS NOT NULL AND qbt_hash <> ''`)
	if err != nil {
		return 0, err
	}
	tracked := map[string]bool{}
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			rows.Close()
			return 0, err
		}
		tracked[strings.TrimSpace(hash)] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	imported := 0
	for hash, torrent := range torrentsByHash {
		if tracked[hash] {
			continue
		}

		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO download_events (user_id, username, filename, torrent_size, qbt_hash, success, user_agent, created_at)
			VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
			userID,
			cleanUsername,
			nullableString(torrent.Name),
			nullableUint64(uint64(max(torrent.Size, 0))),
			hash,
			qbtImportUserAgent,
			time.Unix(int64(torrent.Added_on), 0).UTC(),
		); err != nil {
			return imported, fmt.Errorf("import of %q failed after %d torrents: %w", torrent.Name, imported, err)
		}
		imported++
	}

	return imported, nil
}
