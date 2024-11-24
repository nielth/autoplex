package handlers

import (
	"fmt"
	"net/http"

	"api/services"

	"github.com/gin-gonic/gin"
)

func AuthTokenHandler(c *gin.Context) {
	authURL, respID, clientID := services.InitAuth()
	c.SetCookie("identifier", fmt.Sprintf("%d", respID), 120, "/", "localhost", false, true)
	c.SetCookie("client_identifier", clientID, 120, "/", "localhost", false, true)
	c.JSON(200, gin.H{"url": authURL})
}

func CallbackHandler(c *gin.Context) {
	respID, _ := c.Cookie("identifier")
	clientID, _ := c.Cookie("client_identifier")
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
	c.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"logged_in_as": username})
}

func ProtectedHandler(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(200, gin.H{"message": username})
}
