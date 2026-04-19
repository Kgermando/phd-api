package producers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/controllers/dashboard"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// MiniStats structure for producer quick stats
type MiniStats struct {
	Total       int `json:"total"`
	Eligible    int `json:"eligible"`
	NonEligible int `json:"non_eligible"`
	Femmes      int `json:"femmes"`
}

// ProducerWithScore wraps a producer with its computed total score
type ProducerWithScore struct {
	models.Producer
	TotalScore float64 `json:"total_score"`
}

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

	// // Validation
	// errors := utils.ValidateStruct(producer)
	// if errors != nil {
	// 	return c.Status(400).JSON(errors)
	// }

	// Générer UUID
	producer.UUID = utils.GenerateUUID()

	// Assigner user_uuid depuis le token JWT
	producer.UserUUID = c.Locals("user_uuid").(string)

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

// GetProducerStats retourne les statistiques rapides des producteurs
func GetProducerStats(c *fiber.Ctx) error {
	db := database.DB

	// Get current user UUID from context
	currentUserUUID := c.Locals("user_uuid").(string)

	// Get current user to check role
	var currentUser models.User
	if err := db.First(&currentUser, "uuid = ?", currentUserUUID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch current user",
		})
	}

	var producers []models.Producer
	query := db.Preload("Champs").Preload("Scores")

	// If current user is a Producteur, only show their producers
	if currentUser.Role == "Producteur" {
		query = query.Where("user_uuid = ?", currentUserUUID)
	}

	if err := query.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	// Calculate stats
	stats := MiniStats{
		Total:       0,
		Eligible:    0,
		NonEligible: 0,
		Femmes:      0,
	}

	stats.Total = len(producers)

	for _, producer := range producers {
		// Count women
		if producer.Sexe == "femme" {
			stats.Femmes++
		}

		// Calculate score using dashboard scoring function
		scoreResult := dashboard.ScoreProducer(producer)

		// Count eligible/non-eligible
		if scoreResult.Total >= 60 {
			stats.Eligible++
		} else {
			stats.NonEligible++
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"stats":  stats,
	})
}

// GetPaginatedProducers récupère les producteurs avec pagination
func GetPaginatedProducers(c *fiber.Ctx) error {
	db := database.DB

	// Get current user UUID from context
	currentUserUUID := c.Locals("user_uuid").(string)

	// Get current user to check role
	var currentUser models.User
	if err := db.First(&currentUser, "uuid = ?", currentUserUUID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch current user",
		})
	}

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
	province := c.Query("province", "")
	territoire := c.Query("territoire", "")
	village := c.Query("village", "")
	userUUID := c.Query("user_uuid", "")
	zone := c.Query("zone", "")

	var producers []models.Producer
	var totalRecords int64

	query := db.Model(&models.Producer{})

	// If current user is a Producteur, only show their producers
	if currentUser.Role == "Producteur" {
		query = query.Where("user_uuid = ?", currentUserUUID)
	} else if userUUID != "" {
		// Only allow filtering by user_uuid if not a Producteur
		query = query.Where("user_uuid = ?", userUUID)
	}

	// Add filters
	if search != "" {
		query = query.Where("nom ILIKE ? OR telephone ILIKE ? OR village ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if province != "" {
		query = query.Where("province = ?", province)
	}
	if territoire != "" {
		query = query.Where("territoire = ?", territoire)
	}
	if village != "" {
		query = query.Where("village = ?", village)
	}
	if zone != "" {
		query = query.Where("zone = ?", zone)
	}

	// Count total records
	query.Count(&totalRecords)

	query = db.Model(&models.Producer{})

	// If current user is a Producteur, only show their producers
	if currentUser.Role == "Producteur" {
		query = query.Where("user_uuid = ?", currentUserUUID)
	} else if userUUID != "" {
		// Only allow filtering by user_uuid if not a Producteur
		query = query.Where("user_uuid = ?", userUUID)
	}

	if search != "" {
		query = query.Where("nom ILIKE ? OR telephone ILIKE ? OR village ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if province != "" {
		query = query.Where("province = ?", province)
	}
	if territoire != "" {
		query = query.Where("territoire = ?", territoire)
	}
	if village != "" {
		query = query.Where("village = ?", village)
	}
	if zone != "" {
		query = query.Where("zone = ?", zone)
	}

	err = query.Preload("User").Preload("Champs").Preload("Scores").
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

	producersWithScore := make([]ProducerWithScore, len(producers))
	for i, p := range producers {
		scoreResult := dashboard.ScoreProducer(p)
		producersWithScore[i] = ProducerWithScore{
			Producer:   p,
			TotalScore: scoreResult.Total,
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":    "success",
		"total":     totalRecords,
		"page":      page,
		"limit":     limit,
		"producers": producersWithScore,
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
	if err := db.Preload("User").Preload("Champs").Preload("Scores").First(&producer, "uuid = ?", producerUUID).Error; err != nil {
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

	var input models.Producer
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Mise à jour explicite de tous les champs (nécessaire pour les booléens)
	producer.Nom = input.Nom
	producer.Sexe = input.Sexe
	producer.DateNaissance = input.DateNaissance
	producer.Telephone = input.Telephone
	producer.Province = input.Province
	producer.Territoire = input.Territoire
	producer.Village = input.Village
	producer.Groupement = input.Groupement
	// Section 2
	producer.StatutFoncier = input.StatutFoncier
	producer.AnneesExperience = input.AnneesExperience
	producer.MembreCooperative = input.MembreCooperative
	producer.NomCooperative = input.NomCooperative
	// Section 4
	producer.RotationCultures = input.RotationCultures
	producer.UtilisationCompost = input.UtilisationCompost
	producer.SignesDegradation = input.SignesDegradation
	producer.SourceEau = input.SourceEau
	producer.EconomieEau = input.EconomieEau
	producer.ParcelleInondable = input.ParcelleInondable
	producer.UtilisationPesticides = input.UtilisationPesticides
	producer.FormationPesticides = input.FormationPesticides
	producer.PresenceArbres = input.PresenceArbres
	producer.ActiviteDeforestation = input.ActiviteDeforestation
	producer.BaiseFaune = input.BaiseFaune
	// Section 5
	producer.PerteSec = input.PerteSec
	producer.PerteInondation = input.PerteInondation
	producer.PerteVents = input.PerteVents
	producer.StrategiesAdaptation = input.StrategiesAdaptation
	// Section 6
	producer.VarietesCultivees = input.VarietesCultivees
	producer.RendementMoyen = input.RendementMoyen
	producer.CampagnesParAn = input.CampagnesParAn
	// Section 7
	producer.ManqueEau = input.ManqueEau
	producer.IntrantsCouteux = input.IntrantsCouteux
	producer.AccesCredit = input.AccesCredit
	producer.DegradationSols = input.DegradationSols
	producer.ChangementsClimatiques = input.ChangementsClimatiques
	producer.LieuVente = input.LieuVente
	// Section 8
	producer.BesoinsPrioritaires = input.BesoinsPrioritaires
	// Section 9
	producer.Latitude = input.Latitude
	producer.Longitude = input.Longitude
	producer.Zone = input.Zone

	if err := db.Save(&producer).Error; err != nil {
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

	// // Validation
	// errors := utils.ValidateStruct(champ)
	// if errors != nil {
	// 	return c.Status(400).JSON(errors)
	// }

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

// GetProducersByZone récupère les producteurs par village (zone géographique)
func GetProducersByZone(c *fiber.Ctx) error {
	db := database.DB

	village := c.Query("village", "")
	if village == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Village parameter is required",
		})
	}

	var producers []models.Producer
	if err := db.Where("village = ?", village).Preload("User").Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":    "success",
		"village":   village,
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
