package services

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func qbtConfig() (string, string, string, error) {
	qbtURL, err := requiredEnv("QBT_URL")
	if err != nil {
		return "", "", "", err
	}
	qbtUser, err := requiredEnv("QBT_USER")
	if err != nil {
		return "", "", "", err
	}
	qbtPass, err := requiredEnv("QBT_PASS")
	if err != nil {
		return "", "", "", err
	}

	return strings.TrimRight(qbtURL, "/"), qbtUser, qbtPass, nil
}

func qbtLoginHandler() (*string, string, error) {
	qbtURL, qbtUser, qbtPass, err := qbtConfig()
	if err != nil {
		return nil, "", err
	}

	var loginBody bytes.Buffer
	writer := multipart.NewWriter(&loginBody)

	login := map[string]string{"username": qbtUser, "password": qbtPass}

	for k, v := range login {
		if err := writer.WriteField(k, v); err != nil {
			fmt.Printf("Error writing field %s: %s\n", k, err)
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error closing writer: %s\n", err)
		return nil, "", err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/auth/login", qbtURL), &loginBody)
	if err != nil {
		fmt.Printf("Error creating request to qbt: %s\n", err)
		return nil, "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		return nil, "", err
	}

	defer res.Body.Close()

	cookie := res.Header["Set-Cookie"]
	if len(cookie) == 0 {
		fmt.Println("No cookie in qbt")
		return nil, "", fmt.Errorf("No cookie in qbt")
	}

	return &cookie[0], qbtURL, nil
}

type QbtDownloadList struct {
	Added_on      int     `json:"added_on"`
	Amount_left   int     `json:"amount_left"`
	Completion_on int     `json:"completion_on"`
	Hash          string  `json:"hash"`
	Name          string  `json:"name"`
	Num_complete  int     `json:"num_complete"`
	Num_leechs    int     `json:"num_leechs"`
	Num_seeds     int     `json:"num_seeds"`
	Progress      float64 `json:"progress"`
	Size          int     `json:"size"`
	State         string  `json:"state"`
}

func QbtGetDownloadingList() (*[]QbtDownloadList, error) {
	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/torrents/info?filter=downloading", qbtURL), nil)
	if err != nil {
		fmt.Printf("Error creating request to qbt: %s\n", err)
		return nil, err
	}

	req.Header.Add("cookie", *cookie)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	var qbtDownloadList []QbtDownloadList
	if err := json.Unmarshal(body, &qbtDownloadList); err != nil {
		fmt.Println("Can not unmarshal JSON", err)
		return nil, err
	}

	return &qbtDownloadList, nil

}

func QbtGetAllTorrentsByHash() (map[string]QbtDownloadList, error) {
	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/torrents/info?filter=all", qbtURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("cookie", *cookie)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("qbt torrents/info failed with status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var torrents []QbtDownloadList
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, err
	}

	torrentMap := make(map[string]QbtDownloadList, len(torrents))
	for _, torrent := range torrents {
		hash := strings.ToLower(strings.TrimSpace(torrent.Hash))
		if hash == "" {
			continue
		}
		torrentMap[hash] = torrent
	}

	return torrentMap, nil
}

func QbtGetTorrentByHash(qbtHash string) (*QbtDownloadList, error) {
	cleanHash := strings.ToLower(strings.TrimSpace(qbtHash))
	if cleanHash == "" {
		return nil, fmt.Errorf("qbt hash is required")
	}

	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("hashes", cleanHash)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/torrents/info?%s", qbtURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("cookie", *cookie)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("qbt torrent lookup failed with status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var torrents []QbtDownloadList
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, err
	}

	if len(torrents) == 0 {
		return nil, nil
	}

	return &torrents[0], nil
}

func qbtTagFromFid(fid string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_', r == '-':
			return r
		default:
			return -1
		}
	}, strings.TrimSpace(fid))

	if clean == "" {
		return fmt.Sprintf("autoplex_fid_%d", time.Now().UnixNano())
	}

	return "autoplex_fid_" + clean
}

func qbtFetchByTag(cookie string, qbtURL string, tag string) ([]QbtDownloadList, error) {
	params := url.Values{}
	params.Set("tag", tag)
	params.Set("sort", "added_on")
	params.Set("reverse", "true")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v2/torrents/info?%s", qbtURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("cookie", cookie)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("qbt fetch by tag failed with status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var torrents []QbtDownloadList
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, err
	}

	return torrents, nil
}

func qbtResolveHashByTag(cookie string, qbtURL string, tag string) (string, error) {
	for i := 0; i < 5; i++ {
		torrents, err := qbtFetchByTag(cookie, qbtURL, tag)
		if err != nil {
			return "", err
		}

		if len(torrents) > 0 && strings.TrimSpace(torrents[0].Hash) != "" {
			return strings.TrimSpace(torrents[0].Hash), nil
		}

		time.Sleep(400 * time.Millisecond)
	}

	return "", fmt.Errorf("unable to resolve qbt hash after add")
}

// torrentInfoHash returns the lowercase hex SHA1 of the bencoded info dict,
// which is qBittorrent's torrent id for v1 torrents (all TorrentLeech torrents).
// It returns "" when the bytes are not a parseable bencoded torrent so callers
// can fall back to tag resolution.
func torrentInfoHash(data []byte) string {
	if len(data) == 0 || data[0] != 'd' {
		return ""
	}

	pos := 1
	for pos < len(data) && data[pos] != 'e' {
		keyEnd, err := bencodeValueEnd(data, pos)
		if err != nil {
			return ""
		}
		colon := bytes.IndexByte(data[pos:keyEnd], ':')
		if colon < 0 {
			return ""
		}
		key := string(data[pos+colon+1 : keyEnd])

		valEnd, err := bencodeValueEnd(data, keyEnd)
		if err != nil {
			return ""
		}
		if key == "info" {
			sum := sha1.Sum(data[keyEnd:valEnd])
			return hex.EncodeToString(sum[:])
		}
		pos = valEnd
	}

	return ""
}

// bencodeValueEnd returns the index just past the bencoded value starting at pos.
func bencodeValueEnd(data []byte, pos int) (int, error) {
	if pos >= len(data) {
		return 0, fmt.Errorf("unexpected end of bencode")
	}

	switch c := data[pos]; {
	case c == 'i': // i<digits>e
		end := bytes.IndexByte(data[pos:], 'e')
		if end < 0 {
			return 0, fmt.Errorf("unterminated bencode integer")
		}
		return pos + end + 1, nil
	case c == 'l', c == 'd': // list or dict: values until 'e'
		pos++
		for pos < len(data) && data[pos] != 'e' {
			next, err := bencodeValueEnd(data, pos)
			if err != nil {
				return 0, err
			}
			pos = next
		}
		if pos >= len(data) {
			return 0, fmt.Errorf("unterminated bencode container")
		}
		return pos + 1, nil
	case c >= '0' && c <= '9': // <length>:<bytes>
		colon := bytes.IndexByte(data[pos:], ':')
		if colon < 0 {
			return 0, fmt.Errorf("malformed bencode string")
		}
		length, err := strconv.Atoi(string(data[pos : pos+colon]))
		if err != nil || length < 0 {
			return 0, fmt.Errorf("malformed bencode string length")
		}
		end := pos + colon + 1 + length
		if end > len(data) {
			return 0, fmt.Errorf("bencode string out of range")
		}
		return end, nil
	default:
		return 0, fmt.Errorf("invalid bencode token %q", c)
	}
}

func QbtDownload(data *[]byte, category string, fid string, sequential bool, filename string) (string, error) {
	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var torrentData bytes.Buffer
	writer := multipart.NewWriter(&torrentData)
	tag := qbtTagFromFid(fid)

	// Compute the torrent's info-hash locally so we always know qBittorrent's
	// torrent id regardless of how fast (or whether) qbt indexes our tag. This
	// makes the add idempotent: a slow tag lookup or a 409 no longer loses the
	// torrent, which is what previously left installed torrents unrecorded.
	infoHash := torrentInfoHash(*data)

	part, err := writer.CreateFormFile("torrents", "torrentfile.torrent")
	if err != nil {
		return "", err
	}

	// Write the binary data to the part
	_, err = part.Write(*data)
	if err != nil {
		return "", err
	}

	if err := writer.WriteField("savepath", "/downloads/sde/"+category); err != nil {
		return "", err
	}

	if sequential {
		if err := writer.WriteField("sequentialDownload", "true"); err != nil {
			return "", err
		}
	}

	if err := writer.WriteField("tags", tag); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error closing writer: %s\n", err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/torrents/add", qbtURL), &torrentData)
	if err != nil {
		fmt.Printf("Error creating request to qbt: %s\n", err)
		return "", err
	}

	req.Header.Add("cookie", *cookie)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		// qBit 5.x returns 409 when the same info-hash is already loaded.
		// We already know that info-hash, so adopt the existing torrent by hash
		// and (best-effort) re-apply our tag. Fall back to name matching only
		// when we could not compute the hash.
		if infoHash != "" {
			if tagErr := qbtAddTag(*cookie, qbtURL, infoHash, tag); tagErr != nil {
				fmt.Printf("adopted existing qbt torrent %s on 409 but tagging failed: %v\n", infoHash, tagErr)
			}
			return infoHash, nil
		}
		return qbtAdoptExistingByName(*cookie, qbtURL, filename, tag)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("qbt add failed with status %d: %s", res.StatusCode, string(body))
	}

	// The add succeeded. Prefer the locally-computed info-hash (authoritative for
	// v1 torrents) so a slow tag lookup can never make the caller treat a
	// successful add as a failure. Only fall back to tag resolution if we could
	// not parse the torrent.
	if infoHash != "" {
		return infoHash, nil
	}

	qbtHash, err := qbtResolveHashByTag(*cookie, qbtURL, tag)
	if err != nil {
		return "", err
	}

	return qbtHash, nil
}

func normalizeQbtName(name string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimSuffix(name, ".torrent")))
}

func qbtAdoptExistingByName(cookie, qbtURL, filename, tag string) (string, error) {
	target := normalizeQbtName(filename)
	if target == "" {
		return "", fmt.Errorf("qbt add returned 409 but no filename provided to match existing torrent")
	}

	torrentsByHash, err := QbtGetAllTorrentsByHash()
	if err != nil {
		return "", fmt.Errorf("qbt add 409 and lookup for existing torrent failed: %w", err)
	}

	matched := make([]QbtDownloadList, 0)
	for _, torrent := range torrentsByHash {
		if normalizeQbtName(torrent.Name) == target {
			matched = append(matched, torrent)
		}
	}

	switch len(matched) {
	case 0:
		return "", fmt.Errorf("qbt add returned 409 (already present) but no torrent in qbt matches name %q", filename)
	case 1:
		// continue
	default:
		return "", fmt.Errorf("qbt add returned 409 and %d torrents match name %q; adoption aborted", len(matched), filename)
	}

	hash := matched[0].Hash
	if err := qbtAddTag(cookie, qbtURL, hash, tag); err != nil {
		return hash, fmt.Errorf("adopted existing qbt torrent %s but tagging failed: %w", hash, err)
	}

	fmt.Printf("adopted existing qbt torrent name=%q hash=%s tag=%s\n", matched[0].Name, hash, tag)
	return hash, nil
}

func qbtAddTag(cookie, qbtURL, hash, tag string) error {
	formData := url.Values{}
	formData.Set("hashes", hash)
	formData.Set("tags", tag)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/torrents/addTags", qbtURL), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("qbt addTags failed with status %d: %s", res.StatusCode, string(body))
	}
	return nil
}

func QbtDelete(qbtHash string) error {
	cleanHash := strings.TrimSpace(qbtHash)
	if cleanHash == "" {
		return fmt.Errorf("qbt hash is required")
	}

	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		return err
	}

	formData := url.Values{}
	formData.Set("hashes", cleanHash)
	formData.Set("deleteFiles", "true")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/torrents/delete", qbtURL), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("cookie", *cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("qbt delete failed with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}

func QbtPause(qbtHash string) error {
	cleanHash := strings.TrimSpace(qbtHash)
	if cleanHash == "" {
		return fmt.Errorf("qbt hash is required")
	}

	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		return err
	}

	formData := url.Values{}
	formData.Set("hashes", cleanHash)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/torrents/stop", qbtURL), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("cookie", *cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("qbt stop failed with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}
