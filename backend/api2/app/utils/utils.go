package utils

import (
	"crypto/ed25519"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func FetchAndParseXML(url string, result any, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		errCh <- fmt.Errorf("failed to send request to Plex Server (Hidden Plex Server URL)")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errCh <- fmt.Errorf("failed to read response body received from Plex Server (Hidden Plex Server URL)")
		return
	}

	if err := xml.Unmarshal(body, result); err != nil {
		errCh <- fmt.Errorf("failed to parse XML from Plex Server (Hidden Plex Server URL)")
		return
	}
}

var (
	PubKey  ed25519.PublicKey
	PrivKey ed25519.PrivateKey
)

func LoadKeys() error {
	privKeyData, err := os.ReadFile("ed25519_priv.key")
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}
	pubKeyData, err := os.ReadFile("ed25519_pub.key")
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	PrivKey = ed25519.PrivateKey(privKeyData)
	PubKey = ed25519.PublicKey(pubKeyData)
	return nil
}

// func genEd25519(){
// 	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	os.WriteFile("ed25519_pub.key", pubKey, 0600)
// 	os.WriteFile("ed25519_priv.key", privKey, 0600)
// }
