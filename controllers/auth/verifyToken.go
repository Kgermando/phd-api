package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/dentic-support-api/database"
	"github.com/kgermando/dentic-support-api/models"
	"gorm.io/gorm"
)

// VerifyResetToken vérifie la validité d'un token de réinitialisation
func VerifyResetToken(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Token requis",
		})
	}

	// Recherche du token
	passwordReset := &models.PasswordReset{}
	result := database.DB.Where("token = ? AND used = false", token).First(passwordReset)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"valid":   false,
				"message": "Token invalide",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur de base de données",
		})
	}

	// Vérification de l'expiration
	if time.Now().After(passwordReset.ExpirationTime) {
		// Marquer le token comme utilisé
		database.DB.Model(passwordReset).Update("used", true)
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"valid":   false,
			"message": "Token expiré",
		})
	}

	// Recherche de l'employé pour vérifier qu'il existe toujours
	user := &models.Agent{}
	result = database.DB.Where("email = ?", passwordReset.Email).First(user)
	if result.Error != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"valid":   false,
			"message": "Utilisateur non trouvé",
		})
	}

	// Vérification du statut de l'employé
	if !user.Status {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"valid":   false,
			"message": "Compte désactivé",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"valid":   true,
		"email":   passwordReset.Email,
		"message": "Token valide",
	})
}
