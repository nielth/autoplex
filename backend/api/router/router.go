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

	protected := r.Group("/api/protected")
	protected.Use(middlewares.AuthenticateMiddleware)
	protected.GET("/", handlers.ProtectedHandler)

	return r
}
