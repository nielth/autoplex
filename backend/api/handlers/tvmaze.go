package handlers

import (
	"api/services"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type tvInstallRequest struct {
	Quality string `json:"quality"`
}

type tvAutoInstallRequest struct {
	Enabled bool   `json:"enabled"`
	Quality string `json:"quality"`
}

func TvMazeSearchShowsHandler(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))

	results, err := services.TvMazeSearchShows(query)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func TvMazeShowDetailHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	show, err := services.TvMazeGetShow(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	seasons, err := services.TvMazeGetSeasons(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	episodes, err := services.TvMazeGetEpisodes(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	username := c.GetString("username")
	if services.IsTvMazeShowEnded(show.Status) {
		if err := services.DisableTvShowAutoInstallUpcoming(username, showID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	installStatus, err := services.GetTvShowInstallStatus(username, showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"show":          show,
		"seasons":       seasons,
		"episodes":      episodes,
		"installStatus": installStatus,
	})
}

func TvMazeShowInstallStatusHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	show, err := services.TvMazeGetShow(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	username := c.GetString("username")
	if services.IsTvMazeShowEnded(show.Status) {
		if err := services.DisableTvShowAutoInstallUpcoming(username, showID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	status, err := services.GetTvShowInstallStatus(username, showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func TvMazeConfigureAutoInstallHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	var input tvAutoInstallRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	show, err := services.TvMazeGetShow(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	username := c.GetString("username")
	enabled := input.Enabled
	if services.IsTvMazeShowEnded(show.Status) {
		enabled = false
	}

	subscription, err := services.ConfigureTvShowAutoInstall(username, *show, input.Quality, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status, err := services.GetTvShowInstallStatus(username, showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription":  subscription,
		"installStatus": status,
	})
}

func TvMazeConfigurePreferredQualityHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	var input tvInstallRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	show, err := services.TvMazeGetShow(showID)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	username := c.GetString("username")
	subscription, err := services.ConfigureTvShowPreferredQuality(username, *show, input.Quality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status, err := services.GetTvShowInstallStatus(username, showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription":  subscription,
		"installStatus": status,
	})
}

func TvMazeInstallWholeShowHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	var input tvInstallRequest
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.GetString("username")
	result, err := services.QueueWholeShowInstall(username, showID, input.Quality)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, result)
}

func TvMazeInstallSeasonHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	seasonNumber, err := strconv.Atoi(c.Param("season"))
	if err != nil || seasonNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season number"})
		return
	}

	var input tvInstallRequest
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.GetString("username")
	result, err := services.QueueSeasonInstall(username, showID, seasonNumber, input.Quality)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, result)
}

func TvMazeInstallEpisodeHandler(c *gin.Context) {
	showID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || showID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid show id"})
		return
	}

	episodeID, err := strconv.ParseInt(c.Param("episode"), 10, 64)
	if err != nil || episodeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid episode id"})
		return
	}

	var input tvInstallRequest
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.GetString("username")
	result, err := services.QueueEpisodeInstall(username, showID, episodeID, input.Quality)
	if err != nil {
		handleTvMazeError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, result)
}

func handleTvMazeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrTvMazeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tvmaze resource not found"})
	default:
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}
