package middlewares

import (
	"api/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthenticateMiddleware(c *gin.Context) {
	// Retrieve the token from the cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing in cookie"})
		c.Abort()
		return
	}

	// Verify the token
	if token, err := services.VerifyToken(tokenString); err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
		c.Abort()
		return
	} else {
		c.Set("username", token.Claims.(jwt.MapClaims)["username"])
	}

	// Continue with the next middleware or route handler
	c.Next()
}
