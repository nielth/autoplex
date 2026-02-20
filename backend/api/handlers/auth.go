package handlers

import (
	"fmt"
	"net/http"
	"os"

	"api/services"

	"github.com/gin-gonic/gin"
)

func AuthTokenHandler(c *gin.Context) {
	authURL, respID, clientID, err := services.InitAuth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	domain := os.Getenv("DOMAIN")

	c.SetCookie("identifier", fmt.Sprintf("%d", respID), 120, "/", domain, false, true)
	c.SetCookie("client_identifier", clientID, 120, "/", domain, false, true)
	c.JSON(200, gin.H{"url": authURL})
}

func CallbackHandler(c *gin.Context) {
	respID, _ := c.Cookie("identifier")
	clientID, _ := c.Cookie("client_identifier")
	domain := os.Getenv("DOMAIN")
	username, ok := services.RequestAuthToken(respID, clientID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	tokenString, err := services.CreateToken(username)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating token")
		return
	}
	c.SetCookie("token", tokenString, 60*60*24*7, "/", domain, false, true)
	c.JSON(http.StatusOK, gin.H{"logged_in_as": username})
}
