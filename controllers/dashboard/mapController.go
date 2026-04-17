package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
)

// ============ Map Structures ============

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type MapProducer struct {
	Producer *models.Producer `json:"producer"`
	Score    float64          `json:"score"`
	Position LatLng           `json:"position"`
}

type MapData struct {
	Producers  []MapProducer `json:"producers"`
	Total      int           `json:"total"`
	WithCoords int           `json:"with_coords"` // Number of producers with coordinates
}

// ============ Map Endpoints ============

// GetMapData returns producer data formatted for Google Maps
func GetMapData(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	db.Preload("Champs").
		Preload("Scores").
		Find(&producers)

	var mapProducers []MapProducer
	coordCount := 0

	for _, producer := range producers {
		// Calculate score
		scoreResult := ScoreProducer(producer)

		// Check if producer has coordinates
		if producer.Latitude != nil && producer.Longitude != nil {
			coordCount++

			mapProducers = append(mapProducers, MapProducer{
				Producer: &producer,
				Score:    scoreResult.Total,
				Position: LatLng{
					Lat: *producer.Latitude,
					Lng: *producer.Longitude,
				},
			})
		}
	}

	mapData := MapData{
		Producers:  mapProducers,
		Total:      len(producers),
		WithCoords: coordCount,
	}

	return c.JSON(mapData)
}

// GetMapDataByZone returns producer data filtered by zone for Google Maps
func GetMapDataByZone(c *fiber.Ctx) error {
	zone := c.Query("zone", "")
	if zone == "" {
		return c.Status(400).JSON(fiber.Map{"error": "zone parameter is required"})
	}

	db := database.DB

	var producers []models.Producer
	db.Preload("Champs").
		Preload("Scores").
		Where("village = ?", zone).
		Find(&producers)

	var mapProducers []MapProducer
	coordCount := 0

	for _, producer := range producers {
		scoreResult := ScoreProducer(producer)

		if producer.Latitude != nil && producer.Longitude != nil {
			coordCount++

			mapProducers = append(mapProducers, MapProducer{
				Producer: &producer,
				Score:    scoreResult.Total,
				Position: LatLng{
					Lat: *producer.Latitude,
					Lng: *producer.Longitude,
				},
			})
		}
	}

	mapData := MapData{
		Producers:  mapProducers,
		Total:      len(producers),
		WithCoords: coordCount,
	}

	return c.JSON(mapData)
}

// GetMapDataByScore returns producer data filtered by eligibility for Google Maps
func GetMapDataByScore(c *fiber.Ctx) error {
	eligible := c.Query("eligible", "") // "true" or "false"

	db := database.DB

	var producers []models.Producer
	db.Preload("Champs").
		Preload("Scores").
		Find(&producers)

	var mapProducers []MapProducer
	coordCount := 0

	for _, producer := range producers {
		scoreResult := ScoreProducer(producer)

		// Filter by eligibility (>= 60 is eligible)
		if eligible == "true" && scoreResult.Total < 60 {
			continue
		}
		if eligible == "false" && scoreResult.Total >= 60 {
			continue
		}

		if producer.Latitude != nil && producer.Longitude != nil {
			coordCount++

			mapProducers = append(mapProducers, MapProducer{
				Producer: &producer,
				Score:    scoreResult.Total,
				Position: LatLng{
					Lat: *producer.Latitude,
					Lng: *producer.Longitude,
				},
			})
		}
	}

	mapData := MapData{
		Producers:  mapProducers,
		Total:      len(producers),
		WithCoords: coordCount,
	}

	return c.JSON(mapData)
}
