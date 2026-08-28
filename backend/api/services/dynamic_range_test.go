package services

import "testing"

func tvTorrent(name string, seeders int) TlSeriesTorrent {
	return TlSeriesTorrent{
		Fid:        name,
		Filename:   name + ".torrent",
		CategoryID: 32,
		Name:       name,
		Seeders:    seeders,
	}
}

var (
	siloPlain = tvTorrent("Silo.S03E09.2160p.WEB.H265-CAKES", 100)
	siloDV    = tvTorrent("Silo.S03E09.Farewell.2160p.ATVP.WEB-DL.DDP5.1.Atmos.DV.HDR.H.265-FLUX", 10)
	siloHDR   = tvTorrent("Silo.S03E09.Farewell.2160p.ATVP.WEB-DL.DDP5.1.Atmos.HDR.H.265-NTb", 20)
	silo1080  = tvTorrent("Silo.S03E09.1080p.WEB.H264-CAKES", 50)
)

func TestSelectBestTorrentByQualityDynamicRange(t *testing.T) {
	cases := []struct {
		name         string
		torrents     []TlSeriesTorrent
		quality      string
		dynamicRange string
		expected     string
	}{
		{
			name:         "any takes the most seeded release",
			torrents:     []TlSeriesTorrent{siloPlain, siloDV, siloHDR},
			quality:      "2160",
			dynamicRange: "any",
			expected:     siloPlain.Name,
		},
		{
			name:         "dv prefers dolby vision over more seeded plain release",
			torrents:     []TlSeriesTorrent{siloPlain, siloHDR, siloDV},
			quality:      "2160",
			dynamicRange: "dv",
			expected:     siloDV.Name,
		},
		{
			name:         "dv falls back to hdr",
			torrents:     []TlSeriesTorrent{siloPlain, siloHDR},
			quality:      "2160",
			dynamicRange: "dv",
			expected:     siloHDR.Name,
		},
		{
			name:         "dv skips releases without dv or hdr",
			torrents:     []TlSeriesTorrent{siloPlain},
			quality:      "2160",
			dynamicRange: "dv",
			expected:     "",
		},
		{
			name:         "hdr ignores plain releases",
			torrents:     []TlSeriesTorrent{siloPlain, siloHDR},
			quality:      "2160",
			dynamicRange: "hdr",
			expected:     siloHDR.Name,
		},
		{
			name:         "hdr rejects a dv release tagged hdr",
			torrents:     []TlSeriesTorrent{siloPlain, siloDV},
			quality:      "2160",
			dynamicRange: "hdr",
			expected:     "",
		},
		{
			name:         "hdr takes the plain hdr release next to a dv one",
			torrents:     []TlSeriesTorrent{siloDV, siloHDR},
			quality:      "2160",
			dynamicRange: "hdr",
			expected:     siloHDR.Name,
		},
		{
			name:         "dynamic range is ignored at 1080p",
			torrents:     []TlSeriesTorrent{silo1080},
			quality:      "1080",
			dynamicRange: "dv",
			expected:     silo1080.Name,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			selected := SelectBestTorrentByQuality(testCase.torrents, testCase.quality, testCase.dynamicRange)
			if testCase.expected == "" {
				if selected != nil {
					t.Fatalf("expected no torrent, got %q", selected.Name)
				}
				return
			}
			if selected == nil {
				t.Fatalf("expected %q, got no torrent", testCase.expected)
			}
			if selected.Name != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, selected.Name)
			}
		})
	}
}

func TestSelectBestBoxsetTorrentByQualityDynamicRange(t *testing.T) {
	plain := tvTorrent("Silo.S03.2160p.WEB.H265-CAKES", 100)
	dolbyVision := tvTorrent("Silo.S03.2160p.ATVP.WEB-DL.DDP5.1.Atmos.DoVi.HDR.H.265-FLUX", 10)

	selected := SelectBestBoxsetTorrentByQuality([]TlSeriesTorrent{plain, dolbyVision}, "Silo", 3, "2160", "dv")
	if selected == nil || selected.Name != dolbyVision.Name {
		t.Fatalf("expected %q, got %v", dolbyVision.Name, selected)
	}

	selected = SelectBestBoxsetTorrentByQuality([]TlSeriesTorrent{plain}, "Silo", 3, "2160", "hdr")
	if selected != nil {
		t.Fatalf("expected no boxset, got %q", selected.Name)
	}
}

func TestDynamicRangeTokenDetection(t *testing.T) {
	if !HasDolbyVisionToken("Silo.S03E09.2160p.WEB-DL.Dolby.Vision.H.265-FLUX") {
		t.Fatal("expected 'Dolby.Vision' to be detected as dolby vision")
	}
	if HasDolbyVisionToken("Silo.S03E09.2160p.DVDRip.H264-CAKES") {
		t.Fatal("expected 'DVDRip' not to be detected as dolby vision")
	}
	if !HasHdrToken("Silo.S03E09.2160p.WEB-DL.HDR10+.H.265-FLUX") {
		t.Fatal("expected 'HDR10+' to be detected as hdr")
	}
	if HasHdrToken("Silo.S03E09.2160p.WEB.H265-CAKES") {
		t.Fatal("expected plain release not to be detected as hdr")
	}
}
