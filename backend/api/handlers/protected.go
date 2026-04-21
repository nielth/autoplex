package handlers

import (
	"api/models"
	"api/services"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ProtectedHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")
	c.JSON(200, gin.H{"logged_in_as": username, "is_admin": isAdmin})
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

	if markErr := services.MarkSearchResultsWithDownloaded(resp); markErr != nil {
		log.Printf("failed to mark downloaded search results: %v", markErr)
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
		if logErr := services.LogDownloadEvent(username, data, "", false, "invalid download payload", c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isDownloaded, err := services.IsFidAlreadyDownloaded(data.Fid)
	if err != nil {
		if logErr := services.LogDownloadEvent(username, data, "", false, "failed duplicate download check", c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate download request"})
		return
	}

	if isDownloaded {
		if logErr := services.LogDownloadEvent(username, data, "", false, "torrent already downloaded", c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusConflict, gin.H{"error": "torrent already downloaded"})
		return
	}

	qbtHash, err := services.TlDownloadRequest(data)
	if err != nil {
		if logErr := services.LogDownloadEvent(username, data, "", false, err.Error(), c.ClientIP(), c.Request.UserAgent()); logErr != nil {
			log.Printf("failed to write download event: %v", logErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := services.LogDownloadEvent(username, data, qbtHash, true, "", c.ClientIP(), c.Request.UserAgent()); err != nil {
		log.Printf("failed to write download event: %v", err)
	}

	services.ScheduleAutoPlexScanForDownload(qbtHash, data.CategoryID)

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}

func DownloadListHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")

	downloads, err := services.ListDownloadEvents(username, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"downloads": downloads})
}

func DownloadDeleteHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")

	downloadID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid download id"})
		return
	}

	var input models.DownloadDeleteRequestInput
	if bindErr := c.ShouldBindJSON(&input); bindErr != nil && !errors.Is(bindErr, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErr.Error()})
		return
	}

	action, svcErr := services.DeleteOrRequestDownload(downloadID, username, isAdmin, input.Reason)
	if svcErr != nil {
		switch {
		case errors.Is(svcErr, services.ErrDownloadNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": svcErr.Error()})
		case errors.Is(svcErr, services.ErrDeleteNotAllowed):
			c.JSON(http.StatusForbidden, gin.H{"error": svcErr.Error()})
		case errors.Is(svcErr, services.ErrDownloadAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{"error": svcErr.Error()})
		case errors.Is(svcErr, services.ErrDownloadMissingHash):
			c.JSON(http.StatusConflict, gin.H{"error": svcErr.Error()})
		case errors.Is(svcErr, services.ErrDeleteRequestAlreadyPending):
			c.JSON(http.StatusConflict, gin.H{"error": svcErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": svcErr.Error()})
		}
		return
	}

	if action == services.DownloadDeleteActionRequested {
		c.JSON(http.StatusAccepted, gin.H{"status": action})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": action})
}

func PendingDeleteRequestsHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")

	requests, err := services.ListPendingDeleteRequests(username, isAdmin)
	if err != nil {
		if errors.Is(err, services.ErrAdminRequired) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func DeleteRequestHistoryHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")

	requests, err := services.ListResolvedDeleteRequests(username, isAdmin, 100)
	if err != nil {
		if errors.Is(err, services.ErrAdminRequired) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func ApproveDeleteRequestHandler(c *gin.Context) {
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	if err := services.ApproveDeleteRequest(requestID, username, isAdmin); err != nil {
		switch {
		case errors.Is(err, services.ErrAdminRequired):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDeleteRequestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDeleteRequestNotPending):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDownloadAlreadyDeleted):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDownloadMissingHash):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func PlexScanMoviesAndShowsHandler(c *gin.Context) {
	scannedSections, err := services.TriggerMoviesAndShowsScan()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "scan_triggered",
		"sections": scannedSections,
	})
}

func SystemOverview(c *gin.Context) {
	diskUsage, err := services.DiskUsage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	qbtDownloadingList, err := services.QbtGetDownloadingList()
	if err != nil {
		log.Printf("qbt unavailable, returning empty download list: %v", err)
		qbtDownloadingList = &[]services.QbtDownloadList{}
	}

	resp := map[string]any{"diskUsage": diskUsage, "qbtDownloadingList": qbtDownloadingList}

	c.JSON(http.StatusOK, resp)

}
