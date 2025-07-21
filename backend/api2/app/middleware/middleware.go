package middleware

import (
	"errors"
	"fmt"
	"main/plex"
	"main/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func returnLogin(ctx *gin.Context) {
	_, errIdent := ctx.Cookie("identifier")
	_, errCli := ctx.Cookie("client_identifier")
	authURL, errAuth := ctx.Cookie("authURL")
	if errCli != nil || errIdent != nil || errAuth != nil {
		var (
			respID   int
			clientID string
		)
		domain := os.Getenv("DOMAIN")
		authURL, respID, clientID = plex.InitAuth()
		ctx.SetCookie("identifier", fmt.Sprintf("%d", respID), 120, "/", domain, false, true)
		ctx.SetCookie("client_identifier", clientID, 120, "/", domain, false, true)
		ctx.SetCookie("authURL", authURL, 120, "/", domain, false, true)
	}

	ctx.Header("HX-Redirect", "/")
	ctx.HTML(http.StatusFound, "login/index.html", gin.H{"url": authURL})
}

func CheckAuth(ctx *gin.Context) {
	// Check if Authorization headers exist
	authorization_cookie, err := ctx.Cookie("Authorization")
	if err != nil {
		returnLogin(ctx)
		ctx.Abort()
		return
	}

	// If authorization, parse them
	// `exp` in token gets automatically checked if expired by jwt library
	tokenVerify, err := jwt.Parse(authorization_cookie, func(token *jwt.Token) (any, error) {
		return utils.PubKey, nil
	})

	// Even though there is error in parsing, the claim can still be read (logging)
	if _, ok := tokenVerify.Claims.(jwt.MapClaims); ok && tokenVerify.Valid {
		// fmt.Println(claim)
	} else {
		fmt.Println("Could not read claim from cookie")
	}

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			fmt.Println("Token has expired.")
		case errors.Is(err, jwt.ErrTokenMalformed):
			fmt.Println("Token is malformed.")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			fmt.Println("Invalid signature.")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			fmt.Println("Token not valid yet.")
		default:
			fmt.Printf("Token parse error: %v\n", err)
		}
		ctx.SetCookie("Authorization", "", -1, "/", "", false, true)
		returnLogin(ctx)
		ctx.Abort()
		return
	}

}
