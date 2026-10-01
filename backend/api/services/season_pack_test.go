package services

import "testing"

func TestIsFullSeasonDrop(t *testing.T) {
	episode := func(number int, airdate string) TvMazeEpisode {
		return TvMazeEpisode{ID: int64(number), Season: 1, Number: number, Airdate: airdate}
	}

	cases := []struct {
		name     string
		episodes []TvMazeEpisode
		expected bool
	}{
		{
			name:     "all episodes share one air date",
			episodes: []TvMazeEpisode{episode(1, "2026-10-01"), episode(2, "2026-10-01"), episode(3, "2026-10-01")},
			expected: true,
		},
		{
			name:     "weekly season",
			episodes: []TvMazeEpisode{episode(1, "2026-10-01"), episode(2, "2026-10-08")},
			expected: false,
		},
		{
			name:     "double premiere followed by weekly episodes",
			episodes: []TvMazeEpisode{episode(1, "2026-10-01"), episode(2, "2026-10-01"), episode(3, "2026-10-08")},
			expected: false,
		},
		{
			name:     "single episode",
			episodes: []TvMazeEpisode{episode(1, "2026-10-01")},
			expected: false,
		},
		{
			name:     "unknown air dates",
			episodes: []TvMazeEpisode{episode(1, ""), episode(2, "")},
			expected: false,
		},
		{
			name:     "no episodes",
			episodes: nil,
			expected: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := isFullSeasonDrop(testCase.episodes); got != testCase.expected {
				t.Fatalf("expected %v, got %v", testCase.expected, got)
			}
		})
	}
}

func TestSelectBestSeasonPackTorrent(t *testing.T) {
	plain := tvTorrent("East of Eden 2026 S01 2160p NF WEB-DL DD+ 5 1 Atmos H 265-KRATOS", 185)
	hdr := tvTorrent("East of Eden 2026 S01 HDR 2160p WEB h265-ETHEL", 30)
	dolbyVision := tvTorrent("East of Eden 2026 S01 DV 2160p WEB h265-ETHEL", 10)
	fullHD := tvTorrent("East of Eden 2026 S01 1080p NF WEB-DL DD+ 5 1 Atmos H 264-KRATOS", 300)
	episode := tvTorrent("East of Eden 2026 S01E01 DV 2160p WEB h265-ETHEL", 500)

	cases := []struct {
		name         string
		torrents     []TlSeriesTorrent
		dynamicRange string
		expected     string
	}{
		{
			name:         "dv settles for an untagged pack",
			torrents:     []TlSeriesTorrent{plain, fullHD, episode},
			dynamicRange: "dv",
			expected:     plain.Name,
		},
		{
			name:         "dv still prefers a dolby vision pack",
			torrents:     []TlSeriesTorrent{plain, hdr, dolbyVision},
			dynamicRange: "dv",
			expected:     dolbyVision.Name,
		},
		{
			name:         "dv falls back to an hdr pack before an untagged one",
			torrents:     []TlSeriesTorrent{plain, hdr},
			dynamicRange: "dv",
			expected:     hdr.Name,
		},
		{
			name:         "hdr settles for an untagged pack",
			torrents:     []TlSeriesTorrent{plain, dolbyVision},
			dynamicRange: "hdr",
			expected:     plain.Name,
		},
		{
			name:         "hdr never takes a dolby vision pack",
			torrents:     []TlSeriesTorrent{dolbyVision},
			dynamicRange: "hdr",
			expected:     "",
		},
		{
			name:         "any takes the most seeded pack",
			torrents:     []TlSeriesTorrent{plain, hdr, dolbyVision},
			dynamicRange: "any",
			expected:     plain.Name,
		},
		{
			name:         "single episodes and other qualities are not packs",
			torrents:     []TlSeriesTorrent{fullHD, episode},
			dynamicRange: "dv",
			expected:     "",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			selected := SelectBestSeasonPackTorrent(testCase.torrents, "East of Eden", 1, "2160", testCase.dynamicRange)
			if testCase.expected == "" {
				if selected != nil {
					t.Fatalf("expected no pack, got %q", selected.Name)
				}
				return
			}
			if selected == nil {
				t.Fatalf("expected %q, got no pack", testCase.expected)
			}
			if selected.Name != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, selected.Name)
			}
		})
	}
}
