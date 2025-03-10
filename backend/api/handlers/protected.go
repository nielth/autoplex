package handlers

import (
	"api/models"
	"api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProtectedHandler(c *gin.Context) {
	username := c.MustGet("username")
	c.JSON(200, gin.H{"logged_in_as": username})
}

func TlSearchHandler(c *gin.Context) {
	search := c.Params.ByName("search")
	page := c.Params.ByName("page")
	resp, err := services.TlSearchRequest(search, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp})
	}
	c.JSON(http.StatusOK, resp)
}

func TlDownloadHandler(c *gin.Context) {
	var data models.DownloadData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	services.TlDownloadRequest(data)
}

func BackendInfo(c *gin.Context) {
	diskUsage, err := services.DiskUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	qbtDownloadingList, err := services.QbtGetDownloadingList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	resp := map[string]any{"diskUsage": diskUsage, "qbtDownloadingList": &qbtDownloadingList}

	c.JSON(http.StatusOK, resp)

}
