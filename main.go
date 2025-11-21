package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BVR-INNOVATION-GROUP/strike-force-backend/config"
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {

	app := fiber.New()

	envErr := godotenv.Load()

	if envErr != nil {
		log.Fatal("Failed to load .env")
	}

	DB, DBError := config.ConnectToDB()

	if DBError != nil {
		log.Fatal("Failed to connect to DB : " + DBError.Error())
	}

	user.RegisterRoutes(app, DB)

	apiV1 := app.Group("/api/v1")
	organization.RegisterRoutes(apiV1, DB)

	fmt.Println("server running on port " + os.Getenv("APP_PORT"))
	app.Listen(":" + os.Getenv("APP_PORT"))

}
