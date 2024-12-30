package router

import (
	"api/handlers"
	"api/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://nielth.github.io/autoplex/"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/authToken", handlers.AuthTokenHandler)
	r.GET("/api/callback", handlers.CallbackHandler)

	protected := r.Group("/api")
	protected.Use(middlewares.AuthenticateMiddleware)
	protected.GET("/protected", handlers.ProtectedHandler)
	protected.GET("/search/:search/:page", handlers.TlSearchHandler)
	protected.POST("/download", handlers.TlDownloadHandler)
	protected.GET("/disk", handlers.DiskUsageHandler)

	return r
}
