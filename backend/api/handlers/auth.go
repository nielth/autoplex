package handlers

import (
	"fmt"
	"log"
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
	respID, respErr := c.Cookie("identifier")
	clientID, clientErr := c.Cookie("client_identifier")
	domain := os.Getenv("DOMAIN")
	if respErr != nil || clientErr != nil {
		if err := services.LogLoginEvent(nil, false, "missing auth callback cookies", c.ClientIP(), c.Request.UserAgent()); err != nil {
			log.Printf("failed to write login event: %v", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	user, ok, authErr := services.RequestAuthToken(respID, clientID)
	if !ok {
		reason := "invalid username or password"
		if authErr != nil {
			reason = authErr.Error()
		}
		if err := services.LogLoginEvent(user, false, reason, c.ClientIP(), c.Request.UserAgent()); err != nil {
			log.Printf("failed to write login event: %v", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	tokenString, err := services.CreateToken(user.Username)
	if err != nil {
		if logErr := services.LogLoginEvent(user, false, "failed to create auth token", c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write login event: %v", logErr)
		}
		c.String(http.StatusInternalServerError, "Error creating token")
		return
	}

	if err := services.LogLoginEvent(user, true, "", c.ClientIP(), c.Request.UserAgent()); err != nil {
		log.Printf("failed to write login event: %v", err)
	}

	c.SetCookie("token", tokenString, 60*60*24*7, "/", domain, false, true)
	c.JSON(http.StatusOK, gin.H{"logged_in_as": user.Username})
}
