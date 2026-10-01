package services

import "testing"

const gb = int64(1e9)

func describeMoves(moves []balancerMove) []string {
	described := make([]string, 0, len(moves))
	for _, move := range moves {
		described = append(described, move.Torrent.Name+"->"+move.To)
	}
	return described
}

func TestPlanDiskBalance(t *testing.T) {
	pool := []string{"sdb", "sdc", "sdd"}
	torrent := func(name string, disk string, sizeGB int64, addedOn int64, heavy bool) balancerTorrent {
		return balancerTorrent{Hash: name, Name: name, Disk: disk, SubPath: "movies", Size: sizeGB * gb, AddedOn: addedOn, Heavy: heavy}
	}

	cases := []struct {
		name     string
		torrents []balancerTorrent
		free     map[string]int64
		expected []string
	}{
		{
			name: "heavy goes straight to a pool disk with room",
			torrents: []balancerTorrent{
				torrent("remux", "sde", 60, 100, true),
			},
			free:     map[string]int64{"sde": 1000 * gb, "sdb": 5 * gb, "sdc": 100 * gb, "sdd": 5 * gb},
			expected: []string{"remux->sdc"},
		},
		{
			name: "full pool disk evicts its oldest light torrents first",
			torrents: []balancerTorrent{
				torrent("remux", "sde", 60, 100, true),
				torrent("old-light", "sdb", 30, 1, false),
				torrent("older-light", "sdb", 40, 0, false),
				torrent("old-heavy", "sdb", 70, 0, true),
				torrent("new-light", "sdb", 50, 50, false),
			},
			free:     map[string]int64{"sde": 1000 * gb, "sdb": 10 * gb, "sdc": 5 * gb, "sdd": 5 * gb},
			expected: []string{"older-light->sde", "old-light->sde"},
		},
		{
			name: "newest heavy avoids disks that recently got heavy torrents",
			torrents: []balancerTorrent{
				torrent("remux", "sde", 60, 100, true),
				torrent("recent-heavy-b", "sdb", 60, 90, true),
				torrent("recent-heavy-c", "sdc", 60, 80, true),
				torrent("old-heavy-d", "sdd", 60, 1, true),
			},
			free:     map[string]int64{"sde": 1000 * gb, "sdb": 500 * gb, "sdc": 500 * gb, "sdd": 500 * gb},
			expected: []string{"remux->sdd"},
		},
		{
			name: "no eviction when spare disk cannot take it",
			torrents: []balancerTorrent{
				torrent("remux", "sde", 60, 100, true),
				torrent("light", "sdb", 80, 1, false),
			},
			free:     map[string]int64{"sde": 50 * gb, "sdb": 5 * gb, "sdc": 5 * gb, "sdd": 5 * gb},
			expected: []string{},
		},
		{
			name: "pool disk with room is filled with newest light torrents",
			torrents: []balancerTorrent{
				torrent("old-light", "sde", 10, 1, false),
				torrent("new-light", "sde", 10, 100, false),
				torrent("too-big", "sde", 200, 200, false),
			},
			free:     map[string]int64{"sde": 1000 * gb, "sdb": 35 * gb, "sdc": 5 * gb, "sdd": 5 * gb},
			expected: []string{"new-light->sdb"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			moves := describeMoves(planDiskBalance(tc.torrents, tc.free, "sde", pool, 20*gb, 50))
			if len(moves) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, moves)
			}
			for i := range moves {
				if moves[i] != tc.expected[i] {
					t.Fatalf("expected %v, got %v", tc.expected, moves)
				}
			}
		})
	}
}

func TestSplitQbtSavePath(t *testing.T) {
	disk, subPath, ok := splitQbtSavePath("/downloads/sdb/tvseries/")
	if !ok || disk != "sdb" || subPath != "tvseries" {
		t.Fatalf("got %q %q %t", disk, subPath, ok)
	}
	if _, _, ok := splitQbtSavePath("/data/sdb/movies"); ok {
		t.Fatalf("expected paths outside /downloads to be ignored")
	}
}
