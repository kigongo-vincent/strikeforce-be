package config

import (
	"fmt"
	"os"

	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectToDB() (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	migrationErr := db.AutoMigrate(&user.User{}, &organization.Organization{}, &project.Project{})

	if migrationErr != nil {
		fmt.Println("Small migration issue: [DB HAS DATA]")
	}

	fmt.Println("Connected to DB successfully")

	return db, nil
}
