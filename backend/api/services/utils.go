package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
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

func TlSearchRequest(search string, page string) interface{} {
	cookieFile, err := os.Open("cookie.json")
	if err != nil {
		fmt.Printf("Error opening JSON: %v\n", err)
		return ""
	}
	defer cookieFile.Close()

	byteValue, _ := io.ReadAll(cookieFile)
	var cookieData map[string]string
	if err := json.Unmarshal(byteValue, &cookieData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return ""
	}

	url := "https://www.torrentleech.org/torrents/browse/list/categories/37,43,14,12,47,15,29,26,32,27,44,36/query/" + search + "/orderby/seeders/order/desc/page/" + page

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return ""
	}

	for key, value := range cookieData {
		cookie := &http.Cookie{
			Name:  key,
			Value: value,
		}
		req.AddCookie(cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return ""
	}

	var respJson interface{}
	if err := json.Unmarshal(body, &respJson); err != nil {
		fmt.Printf("Cannot read json: %v\n", err)
		return ""
	}

	return respJson

}
