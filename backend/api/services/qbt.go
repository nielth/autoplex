package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net/http"
)

func qbtLoginHandler() (*string, error) {
	var loginBody bytes.Buffer
	writer := multipart.NewWriter(&loginBody)

	login := map[string]string{"username": "admin", "password": "adminadmin"}

	for k, v := range login {
		if err := writer.WriteField(k, v); err != nil {
			fmt.Printf("Error writing field %s: %s\n", k, err)
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error closing writer: %s\n", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://qbt.internal.nielth.com/api/v2/auth/login", &loginBody)
	if err != nil {
		fmt.Printf("Error creating request to qbt: %s\n", err)
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		return nil, err
	}

	defer res.Body.Close()

	cookie := res.Header["Set-Cookie"]
	if len(cookie) == 0 {
		fmt.Println("No cookie in qbt")
		return nil, fmt.Errorf("No cookie in qbt")
	}

	return &cookie[0], nil
}

func QbtDownload(data *[]byte, category int) (*string, error) {
	cookie, err := qbtLoginHandler()
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

	if err := writer.WriteField("savepath", "/downloads/sdc/movies"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		fmt.Printf("Error closing writer: %s\n", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://qbt.internal.nielth.com/api/v2/torrents/add", &torrentData)
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
