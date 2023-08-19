package db

import (
	"log"

	"github.com/Safwanseban/2fa-go/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := "postgres://safwan:Safwan@123@localhost:5432/jwt"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error %v", err)
	}

	db.AutoMigrate(
		&models.User{},
	)
	return db

}
