package producers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// CreateScore crée un score pour un producteur et calcule automatiquement le total
func CreateScore(c *fiber.Ctx) error {
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

	var score models.Score
	if err := c.BodyParser(&score); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	score.UUID = utils.GenerateUUID()
	score.ProducerUUID = producerUUID

	// Calculer le score total et le statut de recommandation
	score.CalculateScore()

	if err := db.Create(&score).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create score",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Score créé avec succès",
		"data":    score,
	})
}

// GetScoresByProducer récupère tous les scores d'un producteur
func GetScoresByProducer(c *fiber.Ctx) error {
	db := database.DB

	producerUUID := c.Params("uuid")
	if producerUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid producer UUID",
		})
	}

	var scores []models.Score
	if err := db.Where("producer_uuid = ?", producerUUID).
		Order("created_at DESC").
		Find(&scores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch scores",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   scores,
	})
}

// GetScoreByID récupère un score par son UUID
func GetScoreByID(c *fiber.Ctx) error {
	db := database.DB

	scoreUUID := c.Params("scoreUUID")
	if scoreUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid score UUID",
		})
	}

	var score models.Score
	if err := db.Preload("Producer").First(&score, "uuid = ?", scoreUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Score not found",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   score,
	})
}

// UpdateScore met à jour un score et recalcule le total automatiquement
func UpdateScore(c *fiber.Ctx) error {
	db := database.DB

	scoreUUID := c.Params("scoreUUID")
	if scoreUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid score UUID",
		})
	}

	var score models.Score
	if err := db.First(&score, "uuid = ?", scoreUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Score not found",
		})
	}

	var input models.Score
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Mettre à jour les critères
	score.SuperficieCultivee = input.SuperficieCultivee
	score.ExperienceRiziculture = input.ExperienceRiziculture
	score.StatutFoncierSecurise = input.StatutFoncierSecurise
	score.AccesEau = input.AccesEau
	score.RespectItinerairesTechniques = input.RespectItinerairesTechniques
	score.PratiquesEnvironnementales = input.PratiquesEnvironnementales
	score.VulnerabiliteClimatique = input.VulnerabiliteClimatique
	score.OrganisationCooperative = input.OrganisationCooperative
	score.CapaciteProduction = input.CapaciteProduction
	score.MotivationEngagement = input.MotivationEngagement
	score.InclusionSociale = input.InclusionSociale

	// Recalculer le score total et le statut
	score.CalculateScore()

	if err := db.Save(&score).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update score",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Score mis à jour avec succès",
		"data":    score,
	})
}

// DeleteScore supprime un score
func DeleteScore(c *fiber.Ctx) error {
	db := database.DB

	scoreUUID := c.Params("scoreUUID")
	if scoreUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid score UUID",
		})
	}

	var score models.Score
	if err := db.First(&score, "uuid = ?", scoreUUID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Score not found",
		})
	}

	if err := db.Delete(&score).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete score",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Score supprimé avec succès",
	})
}

// GetRecommendedProducers récupère les producteurs avec un score >= 60
func GetRecommendedProducers(c *fiber.Ctx) error {
	db := database.DB

	var scores []models.Score
	if err := db.Where("recommande = ? AND score_total >= ?", true, 60).
		Preload("Producer").
		Order("score_total DESC").
		Find(&scores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch recommended producers",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"seuil":  60,
		"count":  len(scores),
		"data":   scores,
	})
}
