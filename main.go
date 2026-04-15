package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/routes"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {

	database.Connect()

	app := fiber.New()

	// Initialize default config
	app.Use(logger.New())

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://dentic-support.app, https://f005.backblazeb2.com/, http://localhost:4200",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Expires, Cache-Control, Pragma",
		AllowCredentials: true,
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
	}))

	// Configuration pour servir les fichiers statiques
	app.Static("/uploads", "./uploads")

	routes.Setup(app)

	log.Fatal(app.Listen(getPort()))

}