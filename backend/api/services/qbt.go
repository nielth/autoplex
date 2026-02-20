package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
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

func QbtDownload(data *[]byte, category string) (*string, error) {
	cookie, qbtURL, err := qbtLoginHandler()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var torrentData bytes.Buffer
	writer := multipart.NewWriter(&torrentData)

	part, err := writer.CreateFormFile("torrents", "torrentfile.torrent")
	if err != nil {
		return nil, err
	}

	// Write the binary data to the part
	_, err = part.Write(*data)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteField("savepath", "/downloads/sdc/"+category); err != nil {
		return nil, err
	}

	if err := writer.WriteField("sequentialDownload", "true"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error closing writer: %s\n", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v2/torrents/add", qbtURL), &torrentData)
	if err != nil {
		fmt.Printf("Error creating request to qbt: %s\n", err)
		return nil, err
	}

	req.Header.Add("cookie", *cookie)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		return nil, err
	}

	defer res.Body.Close()

	return nil, nil
}
