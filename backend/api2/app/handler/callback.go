package handler

import (
	"fmt"
	"log"
	_db "main/db"
	"main/plex"
	"main/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) Callback(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "callback/index.html", nil)
}

func (h *Handler) CallbackStatus(ctx *gin.Context) {
	respID, errIdent := ctx.Cookie("identifier")
	clientID, errCli := ctx.Cookie("client_identifier")
	if errCli != nil || errIdent != nil {
		ctx.HTML(http.StatusOK, "callback/error.html", gin.H{"error": "Cannot read cookies from client"})
		return
	}

	// Reset cookies once callback has been called
	ctx.SetCookie("identifier", "", -1, "/", "", false, true)
	ctx.SetCookie("client_identifier", "", -1, "/", "", false, true)
	ctx.SetCookie("authURL", "", -1, "/", "", false, true)

	username, userInfo, err := plex.RequestAuthToken(respID, clientID)
	if err != nil {
		fmt.Println(err)
		ctx.HTML(http.StatusOK, "callback/error.html", gin.H{"error": err})
		return
	}

	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"usr": username,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString(utils.PrivKey)
	if err != nil {
		fmt.Println("Internal Error: Cannot sign key", err)
		ctx.HTML(http.StatusOK, "callback/error.html", gin.H{"error": fmt.Sprintf("Internal Error: Cannot sign key: %s", err)})
		return
	}

	_, err = _db.AddUser(h.DB, username, userInfo)
	if err != nil {
		log.Fatalf("Error adding user: %v", err)
	}

	err = _db.UserLogin(h.DB, username, tokenString)
	if err != nil {
		fmt.Println("Failed to append login try to database", err)
		ctx.HTML(http.StatusOK, "callback/error.html", gin.H{"error": err})
		return
	}

	ctx.SetCookie("Authorization", tokenString, int(time.Hour*24*30), "", "", false, true)
	ctx.Header("HX-Redirect", "/")
	ctx.Status(http.StatusOK)
}
