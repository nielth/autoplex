package services

import (
	"log"
	"strings"
	"sync"
	"time"
)

const (
	autoPlexScanMonitorWindow     = 5 * time.Minute
	autoPlexScanPollInterval      = 15 * time.Second
	autoPlexScanProgressThreshold = 0.01
)

var activeAutoPlexScanMonitors sync.Map

func ScheduleAutoPlexScanForDownload(qbtHash string, categoryID int) {
	if !IsMovieOrTVCategory(categoryID) {
		return
	}

	cleanHash := strings.ToLower(strings.TrimSpace(qbtHash))
	if cleanHash == "" {
		return
	}

	if _, loaded := activeAutoPlexScanMonitors.LoadOrStore(cleanHash, struct{}{}); loaded {
		return
	}

	go monitorDownloadAndTriggerPlexScan(cleanHash)
}

func monitorDownloadAndTriggerPlexScan(qbtHash string) {
	defer activeAutoPlexScanMonitors.Delete(qbtHash)

	deadline := time.Now().Add(autoPlexScanMonitorWindow)
	ticker := time.NewTicker(autoPlexScanPollInterval)
	defer ticker.Stop()

	startedDownloading := false

	for {
		torrent, err := QbtGetTorrentByHash(qbtHash)
		if err != nil {
			log.Printf("auto plex scan: failed to read torrent %s from qBittorrent: %v", qbtHash, err)
		} else if torrent != nil {
			if torrent.Progress > 0 {
				startedDownloading = true
			}

			if torrent.Progress > autoPlexScanProgressThreshold {
				scannedSections, scanErr := TriggerMoviesAndShowsScan()
				if scanErr != nil {
					log.Printf("auto plex scan: failed to trigger scan for torrent %s: %v", qbtHash, scanErr)
				} else {
					log.Printf(
						"auto plex scan: triggered scan for torrent %s after %.2f%% progress (sections: %s)",
						qbtHash,
						torrent.Progress*100,
						strings.Join(scannedSections, ", "),
					)
				}
				return
			}
		}

		if time.Now().After(deadline) {
			if !startedDownloading {
				log.Printf("auto plex scan: torrent %s did not start within 5 minutes, skipping", qbtHash)
			} else {
				log.Printf("auto plex scan: torrent %s stayed at or below 1%% for 5 minutes, skipping", qbtHash)
			}
			return
		}

		<-ticker.C
	}
}
