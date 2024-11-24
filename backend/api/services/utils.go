package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func fetchAndParseXML(url string, result interface{}, wg *sync.WaitGroup, errCh chan<- error) {

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

	defer wg.Done()
}
