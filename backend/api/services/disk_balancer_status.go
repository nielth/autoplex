package services

import (
	"context"
	"log"
	"sync"
	"time"
)

type DiskBalancerDisk struct {
	Name             string `json:"name"`
	Role             string `json:"role"`
	Free             int64  `json:"free"`
	Total            int64  `json:"total"`
	HeavyCount       int    `json:"heavyCount"`
	RecentHeavyCount int    `json:"recentHeavyCount"`
}

type DiskBalancerPlanItem struct {
	Name    string  `json:"name"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Size    int64   `json:"size"`
	Mbps    float64 `json:"mbps"`
	Reason  string  `json:"reason"`
	Pending bool    `json:"pending"`
}

type DiskBalancerTorrent struct {
	Name  string  `json:"name"`
	Disk  string  `json:"disk"`
	Size  int64   `json:"size"`
	Mbps  float64 `json:"mbps"`
	Heavy bool    `json:"heavy"`
}

// DiskBalancerStatus is what the last balancer run saw and planned.
type DiskBalancerStatus struct {
	Running   bool                   `json:"running"`
	CheckedAt time.Time              `json:"checkedAt"`
	Message   string                 `json:"message"`
	HeavyMbps float64                `json:"heavyMbps"`
	Disks     []DiskBalancerDisk     `json:"disks"`
	Plan      []DiskBalancerPlanItem `json:"plan"`
	// The highest bitrate torrents, to see how each one was classified.
	TopBitrates []DiskBalancerTorrent `json:"topBitrates"`
}

type DiskBalancerMoveRecord struct {
	ID           uint64    `json:"id"`
	TorrentName  string    `json:"torrentName"`
	FromDisk     string    `json:"fromDisk"`
	ToDisk       string    `json:"toDisk"`
	Size         int64     `json:"size"`
	Mbps         float64   `json:"mbps"`
	Reason       string    `json:"reason"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"errorMessage"`
	CreatedAt    time.Time `json:"createdAt"`
}

var (
	diskBalancerStatusMu sync.Mutex
	diskBalancerStatus   = DiskBalancerStatus{Message: "not run yet"}
)

func setDiskBalancerStatus(status DiskBalancerStatus) {
	diskBalancerStatusMu.Lock()
	defer diskBalancerStatusMu.Unlock()
	diskBalancerStatus = status
}

func GetDiskBalancerStatus() DiskBalancerStatus {
	diskBalancerStatusMu.Lock()
	defer diskBalancerStatusMu.Unlock()
	return diskBalancerStatus
}

// recordDiskBalancerMove logs a move qbt was asked to do. A failed insert only
// gets logged, it should never stop the balancer.
func recordDiskBalancerMove(move balancerMove, moveErr error) {
	db, err := dbConn()
	if err != nil {
		log.Printf("disk balancer: could not record move: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errorMessage := ""
	if moveErr != nil {
		errorMessage = moveErr.Error()
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO disk_balancer_moves
			(torrent_hash, torrent_name, from_disk, to_disk, size_bytes, mbps, reason, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''))`,
		move.Torrent.Hash, move.Torrent.Name, move.Torrent.Disk, move.To, move.Torrent.Size,
		move.Torrent.Mbps, move.Reason, moveErr == nil, errorMessage,
	); err != nil {
		log.Printf("disk balancer: could not record move: %v", err)
	}
}

func ListDiskBalancerMoves(limit int) ([]DiskBalancerMoveRecord, error) {
	db, err := dbConn()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(
		ctx,
		`SELECT id, torrent_name, from_disk, to_disk, size_bytes, mbps, reason, success, COALESCE(error_message, ''), created_at
		FROM disk_balancer_moves
		ORDER BY created_at DESC, id DESC
		LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	moves := make([]DiskBalancerMoveRecord, 0)
	for rows.Next() {
		var move DiskBalancerMoveRecord
		if err := rows.Scan(
			&move.ID, &move.TorrentName, &move.FromDisk, &move.ToDisk, &move.Size, &move.Mbps,
			&move.Reason, &move.Success, &move.ErrorMessage, &move.CreatedAt,
		); err != nil {
			return nil, err
		}
		moves = append(moves, move)
	}

	return moves, rows.Err()
}
