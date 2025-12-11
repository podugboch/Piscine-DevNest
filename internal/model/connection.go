package model

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connection struct {
	gorm.Model
	UserID     uint `gorm:"not null"`
	ResourceID uint `gorm:"not null"`
	ID         uint `gorm:"primaryKey" json:"id"`
	FromID     uint `json:"from_id"`
	ToID       uint `json:"to_id"`
	Accepted   bool `json:"accepted"`
}

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=postgres password=mypassword dbname=piscine_devnest port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	DB = db
	fmt.Println("Connected to PostgreSQL successfully")
}
