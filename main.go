package main

import (
	"log"

	"inventory-api/config"
	"inventory-api/database"
	"inventory-api/routes"
)

// TODO(1): Start Postgres, then `go run .` should connect, migrate and seed.
// TODO(7): Try graceful shutdown (optional) and log useful startup info.
func main() {
	database.Init(config.DSN())

	router := routes.Setup()
	log.Printf("Inventory API listening on http://localhost:%s", config.Port())
	if err := router.Run(":" + config.Port()); err != nil {
		log.Fatal(err)
	}
}
