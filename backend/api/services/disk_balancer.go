package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The disk balancer keeps the "pool" disks (slow to fill, fast enough to stream
// from) full, and keeps the "spare" disk (where new downloads land) as empty as
// possible. High bitrate releases are moved off the spare disk and spread across
// the pool disks so the newest (most likely watched) ones never share a disk.
//
// Every move is a qBittorrent setLocation, so torrents keep seeding and qbt does
// the copying. When a pool disk is full, its oldest low-bitrate torrents are
// first moved to the spare disk to make room.
//
// Disks are seen by qbt as /downloads/<disk>/... and mounted read-only into
// this container as /<disk> (only used to read free space).

const (
	diskBalancerInterval   = 10 * time.Minute
	diskBalancerQbtRoot    = "/downloads"
	diskBalancerMovieHours = 2.0
	diskBalancerEpisodeMin = 50.0
	// Shortest runtimes we expect, used to skip probing torrents that are too
	// small to be heavy.
	diskBalancerShortestMovieMin   = 60.0
	diskBalancerShortestEpisodeMin = 20.0
)

var (
	diskBalancerOnce        sync.Once
	diskBalancerScanPending bool
	// Estimated bitrate (Mbps) per torrent hash. Torrent contents never change,
	// so this only grows and saves a files call per torrent per run.
	diskBalancerBitrateCache = map[string]float64{}
)

var diskBalancerSettledStates = []string{"uploading", "stalledUP", "pausedUP", "stoppedUP", "queuedUP", "forcedUP"}

type diskBalancerConfig struct {
	enabled   bool
	dryRun    bool
	spare     string
	pool      []string
	minFree   int64
	heavyMbps float64
	hotWindow time.Duration
}

type balancerTorrent struct {
	Hash    string
	Name    string
	Disk    string
	SubPath string // save path below the disk, e.g. "movies"
	Size    int64
	AddedOn int64
	Heavy   bool
}

type balancerMove struct {
	Torrent balancerTorrent
	To      string
	Reason  string
}

func diskBalancerSpareDisk() string {
	if disk := strings.TrimSpace(os.Getenv("DISK_BALANCER_SPARE_DISK")); disk != "" {
		return disk
	}
	return "sde"
}

func qbtDiskPath(disk string, subPath string) string {
	return path.Join(diskBalancerQbtRoot, disk, subPath)
}

// splitQbtSavePath turns "/downloads/sdb/movies" into ("sdb", "movies").
func splitQbtSavePath(savePath string) (string, string, bool) {
	rest, ok := strings.CutPrefix(path.Clean(savePath), diskBalancerQbtRoot+"/")
	if !ok || rest == "" {
		return "", "", false
	}
	disk, subPath, _ := strings.Cut(rest, "/")
	return disk, subPath, true
}

func loadDiskBalancerConfig() diskBalancerConfig {
	cfg := diskBalancerConfig{
		enabled:   envBool("DISK_BALANCER_ENABLED"),
		dryRun:    envBool("DISK_BALANCER_DRY_RUN"),
		spare:     diskBalancerSpareDisk(),
		minFree:   int64(envFloat("DISK_BALANCER_MIN_FREE_GB", 20) * 1e9),
		heavyMbps: envFloat("DISK_BALANCER_HEAVY_MBPS", 40),
		hotWindow: time.Duration(envFloat("DISK_BALANCER_HOT_DAYS", 90)*24) * time.Hour,
	}

	for _, disk := range strings.Split(os.Getenv("DISK_BALANCER_POOL_DISKS"), ",") {
		disk = strings.TrimSpace(disk)
		if disk != "" && disk != cfg.spare {
			cfg.pool = append(cfg.pool, disk)
		}
	}

	return cfg
}

func envBool(key string) bool {
	value, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(key)))
	return value
}

func envFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(key)), 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func StartDiskBalancerWorker() {
	cfg := loadDiskBalancerConfig()
	if !cfg.enabled {
		return
	}
	if len(cfg.pool) == 0 {
		log.Printf("disk balancer: DISK_BALANCER_POOL_DISKS is empty, not starting")
		return
	}

	diskBalancerOnce.Do(func() {
		log.Printf("disk balancer: spare=%s pool=%s dryRun=%t", cfg.spare, strings.Join(cfg.pool, ","), cfg.dryRun)
		go func() {
			for {
				if err := RunDiskBalancer(cfg); err != nil {
					log.Printf("disk balancer: %v", err)
				}
				time.Sleep(diskBalancerInterval)
			}
		}()
	})
}

func RunDiskBalancer(cfg diskBalancerConfig) error {
	torrentsByHash, err := QbtGetAllTorrentsByHash()
	if err != nil {
		return err
	}

	disks := append([]string{cfg.spare}, cfg.pool...)
	torrents := make([]balancerTorrent, 0, len(torrentsByHash))
	for _, torrent := range torrentsByHash {
		// One move at a time: free space is only meaningful once qbt is done.
		if torrent.State == "moving" {
			log.Printf("disk balancer: %q is still moving, waiting", torrent.Name)
			return nil
		}
		if torrent.Progress < 1 || !slices.Contains(diskBalancerSettledStates, torrent.State) {
			continue
		}
		disk, subPath, ok := splitQbtSavePath(torrent.SavePath)
		if !ok || !slices.Contains(disks, disk) {
			continue
		}
		torrents = append(torrents, balancerTorrent{
			Hash:    torrent.Hash,
			Name:    torrent.Name,
			Disk:    disk,
			SubPath: subPath,
			Size:    int64(torrent.Size),
			AddedOn: int64(torrent.Added_on),
		})
	}

	if diskBalancerScanPending && !cfg.dryRun {
		if _, err := TriggerMoviesAndShowsScan(); err != nil {
			log.Printf("disk balancer: plex scan after move failed: %v", err)
		} else {
			diskBalancerScanPending = false
		}
	}

	sessions, err := PlexActiveSessionCount()
	if err != nil {
		return fmt.Errorf("could not read plex sessions, skipping: %w", err)
	}
	if sessions > 0 {
		log.Printf("disk balancer: %d plex stream(s) active, not moving anything", sessions)
		return nil
	}

	free := make(map[string]int64, len(disks))
	for _, disk := range disks {
		var stat syscall.Statfs_t
		if err := syscall.Statfs("/"+disk, &stat); err != nil {
			return fmt.Errorf("could not read free space of /%s (is it mounted?): %w", disk, err)
		}
		free[disk] = int64(stat.Bavail) * int64(stat.Bsize)
	}

	for i := range torrents {
		mbps, err := estimateTorrentMbps(torrents[i], cfg.heavyMbps)
		if err != nil {
			log.Printf("disk balancer: could not estimate bitrate of %q: %v", torrents[i].Name, err)
			continue
		}
		torrents[i].Heavy = mbps >= cfg.heavyMbps
	}

	hotSince := time.Now().Add(-cfg.hotWindow).Unix()
	moves := planDiskBalance(torrents, free, cfg.spare, cfg.pool, cfg.minFree, hotSince)

	for _, move := range moves {
		location := qbtDiskPath(move.To, move.Torrent.SubPath)
		log.Printf(
			"disk balancer: %s %q (%.1f GB) %s -> %s (dryRun=%t)",
			move.Reason, move.Torrent.Name, float64(move.Torrent.Size)/1e9, move.Torrent.Disk, location, cfg.dryRun,
		)
		if cfg.dryRun {
			continue
		}
		if err := QbtSetLocation(move.Torrent.Hash, location); err != nil {
			return fmt.Errorf("move of %q failed: %w", move.Torrent.Name, err)
		}
		diskBalancerScanPending = true
	}

	return nil
}

// estimateTorrentMbps returns the bitrate of the biggest video file, read with
// ffprobe. If probing fails it guesses from the file size, assuming a movie runs
// 2 hours and an episode 50 minutes. Torrents too small to reach heavyMbps even
// at the shortest plausible runtime skip the file list and the probe.
func estimateTorrentMbps(torrent balancerTorrent, heavyMbps float64) (float64, error) {
	isTv := strings.Contains(strings.ToLower(torrent.SubPath), "tv")
	guessSeconds, shortestSeconds := diskBalancerMovieHours*3600, diskBalancerShortestMovieMin*60
	if isTv {
		guessSeconds, shortestSeconds = diskBalancerEpisodeMin*60, diskBalancerShortestEpisodeMin*60
	}

	if maxMbps := float64(torrent.Size) * 8 / shortestSeconds / 1e6; maxMbps < heavyMbps {
		return float64(torrent.Size) * 8 / guessSeconds / 1e6, nil
	}

	if mbps, ok := diskBalancerBitrateCache[torrent.Hash]; ok {
		return mbps, nil
	}

	files, err := QbtGetTorrentFiles(torrent.Hash)
	if err != nil {
		return 0, err
	}

	var biggest QbtTorrentFile
	for _, file := range files {
		if isVideoFile(file.Name) && file.Size > biggest.Size {
			biggest = file
		}
	}
	if biggest.Size == 0 {
		diskBalancerBitrateCache[torrent.Hash] = 0
		return 0, nil
	}

	// qbt file names are relative to the save path; the disk is mounted at /<disk>.
	filePath := path.Join("/", torrent.Disk, torrent.SubPath, biggest.Name)
	mbps, err := ffprobeMbps(filePath)
	if err != nil {
		mbps = float64(biggest.Size) * 8 / guessSeconds / 1e6
		log.Printf("disk balancer: ffprobe failed for %s, guessing %.1f Mbps from size: %v", filePath, mbps, err)
	}

	diskBalancerBitrateCache[torrent.Hash] = mbps
	return mbps, nil
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
		Size     string `json:"size"`
	} `json:"format"`
}

// ffprobeMbps reads the average bitrate of a video file. ffprobe only reads the
// container header, so this is cheap even for large remuxes.
func ffprobeMbps(filePath string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := exec.CommandContext(
		ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration,bit_rate,size", "-of", "json", filePath,
	).Output()
	if err != nil {
		return 0, err
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return 0, err
	}

	if bitRate, err := strconv.ParseFloat(probe.Format.BitRate, 64); err == nil && bitRate > 0 {
		return bitRate / 1e6, nil
	}

	duration, durationErr := strconv.ParseFloat(probe.Format.Duration, 64)
	size, sizeErr := strconv.ParseFloat(probe.Format.Size, 64)
	if durationErr != nil || sizeErr != nil || duration <= 0 {
		return 0, fmt.Errorf("ffprobe returned no bitrate or duration")
	}
	return size * 8 / duration / 1e6, nil
}

func isVideoFile(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".mkv", ".mp4", ".m2ts", ".ts", ".avi", ".m4v":
		return true
	}
	return false
}

// planDiskBalance decides the next moves. It is a pure function so the
// rebalancing rules can be tested without qbt or real disks.
//
//  1. The newest heavy torrent on the spare disk goes to the pool disk with the
//     fewest recent heavy torrents. If that disk is full, it first gets its
//     oldest light torrents moved to the spare disk; the heavy torrent follows
//     on a later run once those moves are done.
//  2. Otherwise, pool disks with room get the newest light torrents from the
//     spare disk that fit.
func planDiskBalance(torrents []balancerTorrent, free map[string]int64, spare string, pool []string, minFree int64, hotSince int64) []balancerMove {
	byDisk := map[string][]balancerTorrent{}
	for _, torrent := range torrents {
		byDisk[torrent.Disk] = append(byDisk[torrent.Disk], torrent)
	}

	onSpare := byDisk[spare]
	sort.Slice(onSpare, func(i, j int) bool { return onSpare[i].AddedOn > onSpare[j].AddedOn })

	targets := rankPoolDisks(byDisk, free, pool, hotSince)

	for _, candidate := range onSpare {
		if !candidate.Heavy {
			continue
		}
		for _, target := range targets {
			need := candidate.Size + minFree - free[target]
			if need <= 0 {
				return []balancerMove{{Torrent: candidate, To: target, Reason: "spread heavy"}}
			}

			evictions := pickEvictions(byDisk[target], need)
			if evictions == nil {
				continue
			}
			var evictBytes int64
			for _, torrent := range evictions {
				evictBytes += torrent.Size
			}
			if free[spare]-evictBytes < minFree {
				continue
			}

			moves := make([]balancerMove, 0, len(evictions))
			for _, torrent := range evictions {
				moves = append(moves, balancerMove{Torrent: torrent, To: spare, Reason: "make room for " + candidate.Name})
			}
			return moves
		}
	}

	moves := []balancerMove{}
	used := map[string]bool{}
	for _, target := range pool {
		room := free[target] - minFree
		for _, torrent := range onSpare {
			if torrent.Heavy || used[torrent.Hash] || torrent.Size > room {
				continue
			}
			moves = append(moves, balancerMove{Torrent: torrent, To: target, Reason: "fill"})
			used[torrent.Hash] = true
			room -= torrent.Size
		}
	}
	return moves
}

// rankPoolDisks orders pool disks by fewest heavy torrents added since hotSince,
// then by the oldest most recent heavy torrent, then by most free space.
func rankPoolDisks(byDisk map[string][]balancerTorrent, free map[string]int64, pool []string, hotSince int64) []string {
	hotCount := map[string]int{}
	lastHeavy := map[string]int64{}
	for _, disk := range pool {
		for _, torrent := range byDisk[disk] {
			if !torrent.Heavy {
				continue
			}
			if torrent.AddedOn >= hotSince {
				hotCount[disk]++
			}
			lastHeavy[disk] = max(lastHeavy[disk], torrent.AddedOn)
		}
	}

	ranked := slices.Clone(pool)
	sort.SliceStable(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		if hotCount[a] != hotCount[b] {
			return hotCount[a] < hotCount[b]
		}
		if lastHeavy[a] != lastHeavy[b] {
			return lastHeavy[a] < lastHeavy[b]
		}
		return free[a] > free[b]
	})
	return ranked
}

// pickEvictions returns the oldest light torrents that together free at least
// need bytes, or nil when the disk does not have enough of them.
func pickEvictions(onDisk []balancerTorrent, need int64) []balancerTorrent {
	light := make([]balancerTorrent, 0, len(onDisk))
	for _, torrent := range onDisk {
		if !torrent.Heavy {
			light = append(light, torrent)
		}
	}
	sort.Slice(light, func(i, j int) bool { return light[i].AddedOn < light[j].AddedOn })

	var freed int64
	for i, torrent := range light {
		freed += torrent.Size
		if freed >= need {
			return light[:i+1]
		}
	}
	return nil
}
