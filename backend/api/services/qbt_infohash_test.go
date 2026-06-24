package services

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

func TestTorrentInfoHash(t *testing.T) {
	// A single 20-byte piece hash keeps the bencode valid and realistic.
	info := "d6:lengthi12e4:name4:test12:piece lengthi16384e6:pieces20:01234567890123456789e"
	want := sha1.Sum([]byte(info))
	wantHex := hex.EncodeToString(want[:])

	cases := map[string]string{
		"info after announce":  "d8:announce18:http://tracker/abc4:info" + info + "e",
		"info first":           "d4:info" + info + "8:announce18:http://tracker/abce",
		"with integer sibling": "d4:info" + info + "13:creation datei1700000000ee",
	}

	for name, torrent := range cases {
		t.Run(name, func(t *testing.T) {
			got := torrentInfoHash([]byte(torrent))
			if got != wantHex {
				t.Fatalf("info hash mismatch: got %q want %q", got, wantHex)
			}
		})
	}
}

func TestTorrentInfoHashRejectsGarbage(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":      {},
		"not a dict": []byte("l4:infoe"),
		"truncated":  []byte("d4:info" + "d6:lengthi12"),
		"no info":    []byte("d8:announce3:abce"),
	} {
		t.Run(name, func(t *testing.T) {
			if got := torrentInfoHash(data); got != "" {
				t.Fatalf("expected empty hash for %s, got %q", name, got)
			}
		})
	}
}
