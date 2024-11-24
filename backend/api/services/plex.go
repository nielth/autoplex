package services

import (
	"api/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var plexToken string = GetPlexToken()

func GetPlexToken() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	fmt.Println(os.Getenv("PLEX_TOKEN"))
	return os.Getenv("PLEX_TOKEN")
}

func InitAuth() (string, int, string) {
	id := uuid.New().String()
	header := ReturnPlexAuthPayload(id)
	client := http.Client{}

	req, _ := http.NewRequest("POST", "https://plex.tv/api/v2/pins.json?strong=true", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Could not send auth initAuth")
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	code := result["code"].(string)
	respID := int(result["id"].(float64))
	authURL := fmt.Sprintf("https://app.plex.tv/auth#!?clientID=%s&code=%s", id, code)

	return authURL, respID, id
}

func RequestAuthToken(respID string, clientID string) (string, bool) {
	header := ReturnPlexAuthPayload(clientID)
	tokenUrl := fmt.Sprintf("https://plex.tv/api/v2/pins/%s", respID)

	client := http.Client{}

	req, err := http.NewRequest("GET", tokenUrl, nil)

	for key, value := range header {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not retrieve Plex Auth Token")
	}

	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
	}

	defer res.Body.Close()

	token, ok := result["authToken"].(string)
	if !ok {
		fmt.Println("Could not receive 'authToken'")
	}

	url1 := fmt.Sprintf("https://plex.tv/users/account/?X-Plex-Token=%s", token)
	url2 := fmt.Sprintf("http://nielth.com:32444/accounts/?X-Plex-Token=%s", plexToken)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	var mediaContainer models.MediaContainer
	var user models.User

	go fetchAndParseXML(url1, &user, &wg, errCh)
	go fetchAndParseXML(url2, &mediaContainer, &wg, errCh)

	wg.Wait()
	close(errCh)

	for err := range errCh {
		fmt.Printf("Error: %v\n", err)
	}

	var ids []string
	var names []string

	for _, account := range mediaContainer.Accounts {
		ids = append(ids, account.ID)
		names = append(names, account.Name)
	}

	if slices.Contains(names, user.Username) || slices.Contains(ids, user.ID) {
		return user.Username, true
	}
	return "", false
}

func ReturnPlexAuthPayload(client_identifer string) map[string]string {
	header := map[string]string{
		"x-plex-product":           "plex auth app (autoplex)",
		"x-plex-version":           "0.69.420",
		"x-plex-device":            "linux",
		"x-plex-platform":          "linux",
		"x-plex-device-name":       "autoplex",
		"x-plex-device-vendor":     "test",
		"x-plex-model":             "",
		"x-plex-client-platform":   "",
		"x-plex-client-identifier": client_identifer,
		"content-type":             "application/json",
		"Accept":                   "application/json",
	}

	return header

}
