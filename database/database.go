package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"inventory-api/models"
)

// DB is the shared GORM handle used by the handlers.
var DB *gorm.DB

// Init connects to PostgreSQL, migrates the schema and seeds initial data.
//
// TODO(1): Handle the AutoMigrate error instead of ignoring it.
// TODO(1): (Optional) Configure the connection pool / GORM logger.
func Init(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database! ", err)
	}

	DB.AutoMigrate(&models.Item{})
	Seed()
}
