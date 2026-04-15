package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/controllers/auth"
	"github.com/kgermando/phd-api/middlewares"
)

// SetupAuthRoutes configure toutes les routes d'authentification
func SetupAuthRoutes(api fiber.Router) {
	a := api.Group("/auth")

	// Routes publiques (sans authentification)
	a.Post("/register", auth.Register)
	a.Post("/login", auth.Login)
	a.Post("/forgot-password", auth.Forgot)
	a.Get("/verify-reset-token/:token", auth.VerifyResetToken)
	a.Post("/reset/:token", auth.ResetPassword)

	// Routes protégées — nécessitent un token JWT valide
	ap := a.Group("", middlewares.IsAuthenticated)
	ap.Get("/agent", auth.AuthAgent)
	ap.Put("/profil/info", auth.UpdateInfo)
	ap.Put("/change-password", auth.ChangePassword)
	ap.Post("/logout", auth.Logout)
}
