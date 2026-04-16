package auth

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

func Register(c *fiber.Ctx) error {

	nu := new(models.User)

	if err := c.BodyParser(&nu); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if nu.Password != nu.PasswordConfirm {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "passwords do not match",
		})
	}

	u := &models.User{
		UUID:       utils.GenerateUUID(),
		Fullname:   nu.Fullname,
		Email:      nu.Email,
		Telephone:  nu.Telephone,
		Role:       nu.Role,
		Permission: nu.Permission,
		Status:     nu.Status,
	}

	u.SetPassword(nu.Password)

	if err := utils.ValidateStruct(*u); err != nil {
		c.Status(400)
		return c.JSON(err)
	}

	database.DB.Create(u)

	return c.JSON(fiber.Map{
		"message": "agent account created",
		"data":    u,
	})
}

func Login(c *fiber.Ctx) error {

	lu := new(models.Login)

	if err := c.BodyParser(&lu); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := utils.ValidateStruct(*lu); err != nil {
		c.Status(400)
		return c.JSON(err)
	}

	u := &models.User{}

	result := database.DB.Where("email = ? OR telephone = ?", lu.Identifier, lu.Identifier).
		First(&u)

	if result.Error != nil {
		c.Status(404)
		return c.JSON(fiber.Map{
			"message": "invalid email or telephone 😰",
		})
	}

	if err := u.ComparePassword(lu.Password); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "mot de passe incorrect! 😰",
		})
	}

	if !u.Status {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "votre compte est désactivé. Contactez l'administrateur 😰",
		})
	}

	token, err := utils.GenerateJwt(u.UUID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    token,
	})

}

func AuthAgent(c *fiber.Ctx) error {

	token := c.Query("token")

	fmt.Println("token", token)

	// cookie := c.Cookies("token")
	agentUUID, _ := utils.VerifyJwt(token)

	fmt.Println("agentUUID", agentUUID)

	u := models.User{}

	database.DB.Where("uuid = ?", agentUUID).First(&u)
	r := &models.UserResponse{
		UUID:       u.UUID,
		Fullname:   u.Fullname,
		Email:      u.Email,
		Telephone:  u.Telephone,
		Role:       u.Role,
		Permission: u.Permission,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
	return c.JSON(r)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // 1 day ,
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
		"Logout":  "success",
	})

}

// User bioprofile
func UpdateInfo(c *fiber.Ctx) error {
	type UpdateDataInput struct {
		Fullname      string `json:"fullname"`
		Email         string `json:"email"`
		Telephone     string `json:"telephone"`
		DirectionUUID string `json:"direction_uuid"`
		BureauUUID    string `json:"bureau_uuid"`
	}
	var updateData UpdateDataInput

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"errors":  err.Error(),
		})
	}

	cookie := c.Cookies("token")

	agentUUID, _ := utils.VerifyJwt(cookie)

	agent := new(models.User)

	db := database.DB

	result := db.Where("uuid = ?", agentUUID).First(&agent)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Agent non trouvé",
		})
	}

	agent.Fullname = updateData.Fullname
	agent.Email = updateData.Email
	agent.Telephone = updateData.Telephone

	db.Save(&agent)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Agent successfully updated",
		"data":    agent,
	})

}

func ChangePassword(c *fiber.Ctx) error {
	type UpdateDataInput struct {
		OldPassword     string `json:"old_password"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
	}
	var updateData UpdateDataInput

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"errors":  err.Error(),
		})
	}

	cookie := c.Cookies("token")

	agentUUID, _ := utils.VerifyJwt(cookie)

	agent := new(models.User)

	result := database.DB.Where("uuid = ?", agentUUID).First(&agent)

	if result.Error != nil {
		c.Status(404)
		return c.JSON(fiber.Map{
			"status":  "error",
			"message": "Agent non trouvé",
		})
	}

	if err := agent.ComparePassword(updateData.OldPassword); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "votre mot de passe n'est pas correct! 😰",
		})
	}

	if updateData.Password != updateData.PasswordConfirm {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "passwords do not match",
		})
	}

	agent.SetPassword(updateData.Password)

	db := database.DB
	db.Save(&agent)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Mot de passe modifié avec succès",
	})

}
