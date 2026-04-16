package database

import (
	"fmt"
	"strconv"

	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	p := utils.Env("DB_PORT")
	port, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		panic("failed to parse database port 😵!")
	}

	DNS := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", utils.Env("DB_HOST"), port, utils.Env("DB_USER"), utils.Env("DB_PASSWORD"), utils.Env("DB_NAME"))
	connection, err := gorm.Open(postgres.Open(DNS), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("Could not connect to the database 😰!")
	}

	DB = connection
	fmt.Println("Database Connected 🎉!")

	// Run migrations
	err = connection.AutoMigrate(
		// Agent & Authentication
		&models.User{},
		&models.PasswordReset{},
		// Producers & Champs
		&models.Producer{},
		&models.Champs{},
		&models.Score{},
	)

	if err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
	} else {
		fmt.Println("Database migrations completed successfully! ✅")
	}

	// Seed default data
	SeedSuperAdmin()
}
