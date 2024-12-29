package services

import (
	"api/models"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sync"
)

func fetchAndParseXML(url string, result interface{}, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		errCh <- fmt.Errorf("failed to fetch URL %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errCh <- fmt.Errorf("failed to read response body from %s: %v", url, err)
		return
	}

	if err := xml.Unmarshal(body, result); err != nil {
		errCh <- fmt.Errorf("failed to parse XML from %s: %v", url, err)
		return
	}
}

func readCookie() (map[string]string, error) {
	cookieFile, err := os.Open("cookie.json")
	if err != nil {
		fmt.Printf("Error opening JSON: %v\n", err)
		return nil, fmt.Errorf("Error opening cookie JSON")
	}
	defer cookieFile.Close()

	byteValue, _ := io.ReadAll(cookieFile)
	var cookieData map[string]string
	if err := json.Unmarshal(byteValue, &cookieData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, fmt.Errorf("Error parsing cookie JSON")
	}

	if len(cookieData) == 0 {
		fmt.Println("Cookie file emtpy")
		return nil, fmt.Errorf("TL Cookie json file emtpy")
	}

	return cookieData, nil
}

func tlGetRequest(url string, ua *string) ([]byte, error) {
	cookieData, err := readCookie()

	if err != nil {
		fmt.Printf("Error reading Cookie: %v\n", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, fmt.Errorf("Error creating request")
	}

	for key, value := range cookieData {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	if ua != nil {
		req.Header.Add("User-Agent", *ua)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, fmt.Errorf("Error sending request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, fmt.Errorf("Error reading request body")
	}

	return body, nil
}

func TlSearchRequest(search string, page string) (map[string]interface{}, error) {
	url := "https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27,44,36/query/" + search + "/orderby/seeders/order/desc/page/" + page

	body, err := tlGetRequest(url, nil)

	var respJson map[string]interface{}
	if err := json.Unmarshal(body, &respJson); err != nil {
		fmt.Printf("Cannot read JSON from request: %v\n", err)
		return nil, fmt.Errorf("Cannot read JSON from request")
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return respJson, nil

}

func TlDownloadRequest(data models.DownloadData) (*string, error) {
	url := "https://www.torrentleech.org/download/" + data.Fid + "/" + data.Filename
	ua := "U_AGENT" // Without this, TL for some reason breaks
	movie_range := []int{8, 9, 11, 37, 43, 14, 12, 13, 47, 15, 29, 36}
	tv_range := []int{26, 32, 27, 44}

	torrent_data, err := tlGetRequest(url, &ua)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var category int

	if slices.Contains(movie_range, data.CategoryID) {
		category = data.CategoryID
	} else if slices.Contains(tv_range, data.CategoryID) {
		category = data.CategoryID
	}

	_, err = QbtDownload(&torrent_data, category)

	if err != nil {
		return nil, err
	}

	return nil, nil

}
