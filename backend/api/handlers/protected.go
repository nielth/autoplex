package handlers

import (
	"api/models"
	"api/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProtectedHandler(c *gin.Context) {
	username := c.MustGet("username")
	c.JSON(200, gin.H{"logged_in_as": username})
}

func TlSearchHandler(c *gin.Context) {
	username := c.GetString("username")
	search := c.Params.ByName("search")
	page := c.Params.ByName("page")
	resp, err := services.TlSearchRequest(search, page)
	if err != nil {
		if logErr := services.LogSearchEvent(username, search, page, false, err.Error(), c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write search event: %v", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := services.LogSearchEvent(username, search, page, true, "", c.ClientIP(), c.Request.UserAgent()); err != nil {
		log.Printf("failed to write search event: %v", err)
	}

	c.JSON(http.StatusOK, resp)
}

func TlDownloadHandler(c *gin.Context) {
	username := c.GetString("username")
	var data models.DownloadData
	if err := c.ShouldBindJSON(&data); err != nil {
		if logErr := services.LogDownloadEvent(username, data, false, "invalid download payload", c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.TlDownloadRequest(data); err != nil {
		if logErr := services.LogDownloadEvent(username, data, false, err.Error(), c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := services.LogDownloadEvent(username, data, true, "", c.ClientIP(), c.Request.UserAgent()); err != nil {
		log.Printf("failed to write download event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}

func BackendInfo(c *gin.Context) {
	diskUsage, err := services.DiskUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	qbtDownloadingList, err := services.QbtGetDownloadingList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	resp := map[string]any{"diskUsage": diskUsage, "qbtDownloadingList": &qbtDownloadingList}

	c.JSON(http.StatusOK, resp)

}
