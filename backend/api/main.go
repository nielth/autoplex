package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func initAuth() (string, int, string) {
	id := uuid.New()
	client_identifer := id.String()

	header := map[string]string{
		"X-Plex-Product":           "Plex Auth App (Autoplex)",
		"X-Plex-Version":           "0.69.420",
		"X-Plex-Device":            "Linux",
		"X-Plex-Platform":          "Linux",
		"X-Plex-Device-Name":       "Autoplex",
		"X-Plex-Device-Vendor":     "Test",
		"X-Plex-Model":             "",
		"X-Plex-Client-Platform":   "",
		"X-Plex-Client-Identifier": client_identifer,
		"Content-Type":             "application/json",
	}

	client := http.Client{}

	req, err := http.NewRequest("POST", "https://plex.tv/api/v2/pins.json?strong=true", nil)
	if err != nil {
		log.Fatalln(err)
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
	}

	defer resp.Body.Close()

	code, ok := result["code"].(string)
	if !ok {
		fmt.Println("Error: 'code' not found or not the expected type")
	}

	resp_identifier, ok := result["id"].(float64)
	if !ok {
		fmt.Println("Error: 'id' not found or not the expected type")
	}

	auth_url := fmt.Sprintf("https://app.plex.tv/auth#!?clientID=%s&code=%s&forwardUrl=%s", client_identifer, code, "http://localhost:8080")

	return auth_url, int(resp_identifier), client_identifer
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// First step authenticating -> get unique oauth ID and link
	r.GET("/api/authToken", func(c *gin.Context) {
		authURL, respID, clientID := initAuth()
		fmt.Println(string(respID))
		c.SetCookie("identifier", fmt.Sprintf("%d", respID), 120, "/", "localhost", false, true)
		c.SetCookie("client_identifier", clientID, 120, "/", "localhost", false, true)
		c.JSON(200, gin.H{"url": authURL})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run("localhost:8080")
}
