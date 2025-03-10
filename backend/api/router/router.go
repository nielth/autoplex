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
	protected.GET("/disk", handlers.BackendInfo)

	return r
}
