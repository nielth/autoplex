package main

import (
	"api/router"
)

func main() {
	r := router.SetupRouter()
	r.Run("localhost:8080")
}
