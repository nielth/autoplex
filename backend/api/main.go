package main

import (
	"api/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file loaded, using process environment: %v", err)
	}

	r := router.SetupRouter()
	r.Run("0.0.0.0:8080")
}
