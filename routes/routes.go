package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kgermando/phd-api/middlewares"
)

// Setup configure toutes les routes de l'application
func Setup(app *fiber.App) {
	// Groupe API principal avec middleware logger
	api := app.Group("/api", logger.New())

	// Routes publiques (sans authentification)
	SetupAuthRoutes(api)

	// Routes protégées — nécessitent un token JWT valide
	protected := api.Group("", middlewares.IsAuthenticated)
	SetupUserRoutes(protected)
	SetupProducerRoutes(protected)
	SetupDashboardRoutes(protected)

}
