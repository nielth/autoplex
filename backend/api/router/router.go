package router

import (
	"api/handlers"
	"api/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/api/authToken", handlers.AuthTokenHandler)
	r.GET("/api/callback", handlers.CallbackHandler)

	protected := r.Group("/api")
	protected.Use(middlewares.AuthenticateMiddleware)
	protected.GET("/protected", handlers.ProtectedHandler)
	protected.GET("/search/:search/:page", handlers.TlSearchHandler)
	protected.POST("/download", handlers.TlDownloadHandler)
	protected.GET("/downloads", handlers.DownloadListHandler)
	protected.POST("/downloads/:id/delete", handlers.DownloadDeleteHandler)
	protected.GET("/downloads/delete-requests", handlers.PendingDeleteRequestsHandler)
	protected.POST("/downloads/delete-requests/:id/approve", handlers.ApproveDeleteRequestHandler)
	protected.POST("/plex/scan/movies-tv", handlers.PlexScanMoviesAndShowsHandler)
	protected.GET("/disk", handlers.BackendInfo)

	return r
}
