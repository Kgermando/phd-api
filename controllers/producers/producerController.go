package producers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// CreateProducer crée un nouveau producteur avec ses champs
func CreateProducer(c *fiber.Ctx) error {
	db := database.DB

	var producer models.Producer
	if err := c.BodyParser(&producer); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validation
	errors := utils.ValidateStruct(producer)
	if errors != nil {
		return c.Status(400).JSON(errors)
	}

	// Générer UUID
	producer.UUID = utils.GenerateUUID()

	if err := db.Create(&producer).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create producer",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":   "success",
		"message":  "Producer created successfully",
		"producer": producer,
	})
}

// GetPaginatedProducers récupère les producteurs avec pagination
func GetPaginatedProducers(c *fiber.Ctx) error {
	db := database.DB

	// Parse query parameters for pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Parse search query
	search := c.Query("search", "")
	zone := c.Query("zone", "")

	var producers []models.Producer
	var totalRecords int64

	query := db.Model(&models.Producer{})

	// Add filters
	if search != "" {
		query = query.Where("nom ILIKE ? OR telephone ILIKE ? OR village ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if zone != "" {
		query = query.Where("zone = ?", zone)
	}

	// Count total records
	query.Count(&totalRecords)

	query = db.Model(&models.Producer{})
	if search != "" {
		query = query.Where("nom ILIKE ? OR telephone ILIKE ? OR village ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if zone != "" {
		query = query.Where("zone = ?", zone)
	}

	err = query.Preload("Champs").
		Offset(offset).
		Limit(limit).
		Order("producers.updated_at DESC").
		Find(&producers).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":    "success",
		"total":     totalRecords,
		"page":      page,
		"limit":     limit,
		"producers": producers,
	})
}

// GetProducerByID récupère un producteur par son UUID
func GetProducerByID(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	var producer models.Producer
	if err := db.Preload("Champs").First(&producer, "uuid = ?", producerUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Producer not found",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":   "success",
		"producer": producer,
	})
}

// UpdateProducer met à jour un producteur
func UpdateProducer(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	var producer models.Producer
	if err := db.First(&producer, "uuid = ?", producerUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Producer not found",
		})
	}

	var updateData models.Producer
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := db.Model(&producer).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update producer",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":   "success",
		"message":  "Producer updated successfully",
		"producer": producer,
	})
}

// DeleteProducer supprime un producteur
func DeleteProducer(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	var producer models.Producer
	if err := db.First(&producer, "uuid = ?", producerUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Producer not found",
		})
	}

	if err := db.Delete(&producer).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete producer",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Producer deleted successfully",
	})
}

// AddChampToProducer ajoute un champ à un producteur
func AddChampToProducer(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	// Vérifier que le producteur existe
	var producer models.Producer
	if err := db.First(&producer, "uuid = ?", producerUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Producer not found",
		})
	}

	var champ models.Champs
	if err := c.BodyParser(&champ); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validation
	errors := utils.ValidateStruct(champ)
	if errors != nil {
		return c.Status(400).JSON(errors)
	}

	champ.UUID = utils.GenerateUUID()
	champ.ProducerUUID = producerUUID
	if err := db.Create(&champ).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add champ",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Champ added successfully",
		"champ":   champ,
	})
}

// GetProducersByZone récupère les producteurs par zone
func GetProducersByZone(c *fiber.Ctx) error {
	db := database.DB

	zone := c.Query("zone", "")
	if zone == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Zone parameter is required",
		})
	}

	var producers []models.Producer
	if err := db.Where("zone = ?", zone).Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":    "success",
		"zone":      zone,
		"count":     len(producers),
		"producers": producers,
	})
}

// GetProducerChamps récupère tous les champs d'un producteur
func GetProducerChamps(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	var champs []models.Champs
	if err := db.Where("producer_uuid = ?", producerUUID).Find(&champs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch champs",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"champs": champs,
	})
}

// DeleteChamp supprime un champ
func DeleteChamp(c *fiber.Ctx) error {
	db := database.DB

	champUUID := c.Params("champUUID")
	if champUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid champ UUID",
		})
	}

	var champ models.Champs
	if err := db.First(&champ, "uuid = ?", champUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Champ not found",
		})
	}

	if err := db.Delete(&champ).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete champ",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Champ deleted successfully",
	})
}
