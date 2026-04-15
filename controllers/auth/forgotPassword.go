package auth

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/dentic-support-api/database"
	"github.com/kgermando/dentic-support-api/models"
	"github.com/kgermando/dentic-support-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Forgot gère la demande de réinitialisation de mot de passe
func Forgot(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest

	// Parsing des données JSON
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Données JSON invalides",
			"errors":  err.Error(),
		})
	}

	// Validation des données d'entrée
	// if err := utils.ValidateStruct(req); err != nil {
	// 	return c.Status(400).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Données invalides",
	// 		"errors":  err,
	// 	})
	// }

	// Recherche de l'employé dans la base de données
	user := &models.Agent{}
	result := database.DB.Where("email = ?", req.Email).First(user)

	// Si l'employé n'existe pas, on retourne toujours le même message pour éviter l'énumération d'emails
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{
				"status":  "success",
				"message": "Si cette adresse email existe dans notre système, vous recevrez un lien de réinitialisation",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur de base de données",
		})
	}

	// Vérification du statut de l'employé
	if !user.Status {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Si cette adresse email existe dans notre système, vous recevrez un lien de réinitialisation",
		})
	}

	// Invalidation des anciens tokens pour cet email
	database.DB.Model(&models.PasswordReset{}).
		Where("email = ? AND used = false AND expiration_time > ?", req.Email, time.Now()).
		Update("used", true)

	// Génération d'un token sécurisé
	token, err := utils.GenerateSecureToken(32) // 64 caractères hex
	if err != nil {
		log.Printf("Erreur génération token: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur interne du serveur",
		})
	}

	// Création de l'enregistrement de réinitialisation
	passwordReset := &models.PasswordReset{
		UUID:           utils.GenerateUUID(),
		Email:          req.Email,
		Token:          token,
		ExpirationTime: time.Now().Add(time.Hour * 3), // 3 heures
		Used:           false,
		CreatedAt:      time.Now(),
	}

	// Sauvegarde en base de données
	if err := database.DB.Create(passwordReset).Error; err != nil {
		log.Printf("Erreur sauvegarde password reset: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la sauvegarde",
		})
	}

	// Envoi de l'email de réinitialisation
	emailService := utils.NewEmailService()
	userFullName := user.Fullname
	if err := emailService.SendPasswordResetEmail(req.Email, token, userFullName); err != nil {
		log.Printf("Erreur envoi email: %v", err)
		// On supprime le token si l'email n'a pas pu être envoyé
		database.DB.Delete(&passwordReset)
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de l'envoi de l'email",
		})
	}

	// Nettoyage des anciens tokens expirés (asynchrone)
	go cleanupExpiredTokens()

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Si cette adresse email existe dans notre système, vous recevrez un lien de réinitialisation dans quelques minutes",
	})
}

// ResetPassword gère la réinitialisation effective du mot de passe
func ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Token requis",
		})
	}

	var resetData models.Reset
	if err := c.BodyParser(&resetData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Données JSON invalides",
			"errors":  err.Error(),
		})
	}

	// Vérification que les mots de passe correspondent
	if resetData.Password != resetData.PasswordConfirm {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Les mots de passe ne correspondent pas",
		})
	}

	// Validation de la force du mot de passe
	if err := utils.ValidatePassword(resetData.Password); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Recherche du token valide
	passwordReset := &models.PasswordReset{}
	result := database.DB.Where("token = ? AND used = false", token).First(passwordReset)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Token invalide ou expiré",
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
			"message": "Le token a expiré",
		})
	}

	// Recherche de l'employé
	user := &models.Agent{}
	result = database.DB.Where("email = ?", passwordReset.Email).First(user)
	if result.Error != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Utilisateur non trouvé",
		})
	}

	// Début de transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Hachage du nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(resetData.Password), 14)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors du hachage du mot de passe",
		})
	}

	// Mise à jour du mot de passe
	if err := tx.Model(user).Update("password", string(hashedPassword)).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la mise à jour du mot de passe",
		})
	}

	// Marquer le token comme utilisé
	if err := tx.Model(passwordReset).Update("used", true).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la mise à jour du token",
		})
	}

	// Invalidation de tous les autres tokens de cet utilisateur
	tx.Model(&models.PasswordReset{}).
		Where("email = ? AND uuid != ? AND used = false", passwordReset.Email, passwordReset.UUID).
		Update("used", true)

	// Validation de la transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la sauvegarde",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Mot de passe réinitialisé avec succès",
	})
}

// cleanupExpiredTokens supprime les tokens expirés (fonction utilitaire)
func cleanupExpiredTokens() {
	cutoffTime := time.Now().Add(-24 * time.Hour)
	database.DB.Where("expiration_time < ? OR used = true", cutoffTime).
		Delete(&models.PasswordReset{})
}
