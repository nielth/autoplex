package services

import (
	"api/models"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
)

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

// Function to request a generated, temporary Plex oauth link with a unique locally generated UUID4
func InitAuth() (string, int, string, error) {
	localUUID := uuid.New().String()
	header := ReturnPlexAuthPayload(localUUID)
	client := http.Client{}
	forwardUrl, err := requiredEnv("OAUTH_FORWARD_URL")
	if err != nil {
		return "", 0, "", err
	}

	// Request a temporary response ID from plex for oauth authentication where a uuid4 is used to to verify this transaction
	req, err := http.NewRequest("POST", "https://plex.tv/api/v2/pins.json?strong=true", nil)
	if err != nil {
		return "", 0, "", err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, "", fmt.Errorf("could not send auth init request: %w", err)
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, "", fmt.Errorf("could not decode init auth response: %w", err)
	}

	// We us our locally generated uuid4 as well as the returned plex code to create the url to authenticate the user
	returnedPlexCode, ok := result["code"].(string)
	if !ok || returnedPlexCode == "" {
		return "", 0, "", fmt.Errorf("could not read Plex code from init auth response")
	}
	respIDFloat, ok := result["id"].(float64)
	if !ok {
		return "", 0, "", fmt.Errorf("could not read Plex id from init auth response")
	}
	respID := int(respIDFloat)
	authURL := fmt.Sprintf("https://app.plex.tv/auth#!?clientID=%s&code=%s&forwardUrl=%s", localUUID, returnedPlexCode, forwardUrl)

	return authURL, respID, localUUID, nil
}

// Function to check the oauth transaction, if true, check if user is linked to local Plex server
func RequestAuthToken(respID string, clientID string) (*models.User, bool, error) {
	respID = strings.TrimSpace(respID)
	clientID = strings.TrimSpace(clientID)
	if respID == "" || clientID == "" {
		return nil, false, fmt.Errorf("missing auth callback cookies")
	}

	// Link to get authToken from user oauth transaction
	header := ReturnPlexAuthPayload(clientID)
	tokenUrl := fmt.Sprintf("https://plex.tv/api/v2/pins/%s", respID)
	plexToken, err := requiredEnv("PLEX_TOKEN")
	if err != nil {
		return nil, false, err
	}
	plexURL, err := requiredEnv("PLEX_URL")
	if err != nil {
		return nil, false, err
	}
	plexURL = strings.TrimRight(plexURL, "/")

	client := http.Client{}

	req, err := http.NewRequest("GET", tokenUrl, nil)
	if err != nil {
		return nil, false, fmt.Errorf("could not create request for plex auth token: %w", err)
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("could not retrieve plex auth token: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, false, fmt.Errorf("error decoding plex auth response: %w", err)
	}

	token, ok := result["authToken"].(string)
	if !ok || token == "" {
		return nil, false, fmt.Errorf("could not receive authToken from plex")
	}

	// Link to get basic information about user oauth session to check if user is linked to plex
	url1 := fmt.Sprintf("https://plex.tv/users/account/?X-Plex-Token=%s", token)
	// Link to local Plex server to get all users linked in order to validate user login request (might take a while for user to appear)
	url2 := fmt.Sprintf("%s/accounts/?X-Plex-Token=%s", plexURL, plexToken)

	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	var mediaContainer models.MediaContainer
	var user models.User

	go fetchAndParseXML(url1, &user, &wg, errCh)
	go fetchAndParseXML(url2, &mediaContainer, &wg, errCh)

	wg.Wait()
	close(errCh)

	var firstErr error
	for err := range errCh {
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return nil, false, fmt.Errorf("failed to validate plex account")
	}

	var ids []string
	var names []string

	for _, account := range mediaContainer.Accounts {
		ids = append(ids, account.ID)
		names = append(names, account.Name)
	}

	if strings.TrimSpace(user.Username) == "" {
		return nil, false, fmt.Errorf("could not read user profile from plex")
	}

	if slices.Contains(names, user.Username) || slices.Contains(ids, user.ID) {
		return &user, true, nil
	}
	return &user, false, fmt.Errorf("user is not linked to local plex server")
}
