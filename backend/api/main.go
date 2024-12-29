package main

import (
	"api/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := router.SetupRouter()
	r.Run("localhost:8080")
}
