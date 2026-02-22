package main

import (
	"api/router"
	"api/services"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file loaded, using process environment: %v", err)
	}

	if err := services.InitMySQL(); err != nil {
		log.Fatalf("Failed to initialize mysql: %v", err)
	}

	services.StartTvEpisodeAutoInstallWorker()

	r := router.SetupRouter()
	r.Run("0.0.0.0:8080")
}
