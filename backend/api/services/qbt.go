package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	Added_on     int     `json:"added_on"`
	Amount_left  int     `json:"amount_left"`
	Hash         string  `json:"hash"`
	Name         string  `json:"name"`
	Num_complete int     `json:"num_complete"`
	Num_leechs   int     `json:"num_leechs"`
	Num_seeds    int     `json:"num_seeds"`
	Progress     float64 `json:"progress"`
	Size         int     `json:"size"`
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

func QbtDownload(data *[]byte, category string, fid string) (string, error) {
	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var torrentData bytes.Buffer
	writer := multipart.NewWriter(&torrentData)
	tag := qbtTagFromFid(fid)

	part, err := writer.CreateFormFile("torrents", "torrentfile.torrent")
	if err != nil {
		return "", err
	}

	// Write the binary data to the part
	_, err = part.Write(*data)
	if err != nil {
		return "", err
	}

	if err := writer.WriteField("savepath", "/downloads/sdc/"+category); err != nil {
		return "", err
	}

	if err := writer.WriteField("sequentialDownload", "true"); err != nil {
		return "", err
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

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("qbt add failed with status %d: %s", res.StatusCode, string(body))
	}

	qbtHash, err := qbtResolveHashByTag(*cookie, qbtURL, tag)
	if err != nil {
		return "", err
	}

	return qbtHash, nil
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
