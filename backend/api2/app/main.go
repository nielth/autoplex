package main

import (
	"log"
	_db "main/db"
	"main/handler"
	"main/middleware"
	"main/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	if err := utils.LoadKeys(); err != nil {
		log.Fatalf("Error loading ed25519 keys: %v", err)
		return
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return
	}

	db, err := _db.InitDB()
	if err != nil {
		log.Fatalf("Error initiating DB: %v", err)
		return
	}

	defer db.Close()

	handler := &handler.Handler{DB: db}

	router := gin.Default()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/**/*")

	router.GET("/callback", handler.Callback)
	router.GET("/callback/status", handler.CallbackStatus)

	autorized := router.Group("/")
	autorized.Use(middleware.CheckAuth)
	{
		autorized.GET("/", handler.Home)
	}

	router.Run(":8081")
}
