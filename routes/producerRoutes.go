package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/controllers/producers"
)

// SetupProducerRoutes configure les routes producers
func SetupProducerRoutes(api fiber.Router) {
	// ============================================================
	// PRODUCERS ROUTES
	// ============================================================
	prod := api.Group("/producers")

	// Producers CRUD
	prod.Post("/create", producers.CreateProducer)
	prod.Get("/all/paginate", producers.GetPaginatedProducers)
	prod.Get("/get/:uuid", producers.GetProducerByID)
	prod.Put("/update/:uuid", producers.UpdateProducer)
	prod.Delete("/delete/:uuid", producers.DeleteProducer)
	prod.Get("/zone", producers.GetProducersByZone)

	// Champs management
	prod.Post("/:uuid/champs/add", producers.AddChampToProducer)
	prod.Get("/:uuid/champs", producers.GetProducerChamps)
	prod.Delete("/champs/:champUUID", producers.DeleteChamp)
}
