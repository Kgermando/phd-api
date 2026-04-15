package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/controllers/users"
)

// SetupUserRoutes configure les routes users 
func SetupUserRoutes(api fiber.Router) {
	// ============================================================
	// USERS ROUTES
	// ============================================================
	ag := api.Group("/users")
	ag.Get("/all/paginate", users.GetPaginatedUsers)
	ag.Get("/all", users.GetAllUSers) 
	ag.Post("/create", users.CreateUser)
	ag.Get("/get/:uuid", users.GetUser)
	ag.Put("/update/:uuid", users.UpdateUser)
	ag.Delete("/delete/:uuid", users.DeleteUser)

}
