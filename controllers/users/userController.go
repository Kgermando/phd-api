package users

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/kgermando/phd-api/utils"
)

// Paginate
func GetPaginatedUsers(c *fiber.Ctx) error {
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

	var users []models.User
	var totalRecords int64

	// Count total records matching the search query
	db.Model(&models.User{}).
		Where("fullname ILIKE ? OR role ILIKE ?", "%"+search+"%", "%"+search+"%").
		Count(&totalRecords)

	err = db.
		Where("fullname ILIKE ? OR role ILIKE ?", "%"+search+"%", "%"+search+"%").
		Preload("Direction").
		Preload("Bureau").
		Offset(offset).
		Limit(limit).
		Order("users.updated_at DESC").
		Find(&users).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch users",
			"error":   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	//  Prepare pagination metadata
	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	// Return response
	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "users retrieved successfully",
		"data":       users,
		"pagination": pagination,
	})
}

// query all data
func GetAllUSers(c *fiber.Ctx) error {
	db := database.DB
	var users []models.User
	db.Preload("Direction").Preload("Bureau").Find(&users)
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All users",
		"data":    users,
	})
}

// Get one data
func GetUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var user models.User
	db.Where("uuid = ?", uuid).Preload("Direction").Preload("Bureau").First(&user)
	if user.Fullname == "" {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "No user name found",
				"data":    nil,
			},
		)
	}
	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "user found",
			"data":    user,
		},
	)
}

// Create data
func CreateUser(c *fiber.Ctx) error {
	p := &models.User{}

	if err := c.BodyParser(&p); err != nil {
		return err
	}

	if p.Fullname == "" {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Form not complete",
				"data":    nil,
			},
		)
	}

	if p.Password != p.PasswordConfirm {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "passwords do not match",
		})
	}

	user := &models.User{
		Fullname:      p.Fullname,
		Email:         p.Email,
		Telephone:     p.Telephone, 
		Role:          p.Role,
		Permission:    p.Permission,
		Status:        p.Status,
	}

	user.SetPassword(p.Password)

	user.UUID = utils.GenerateUUID()

	database.DB.Create(user)

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "user Created success",
			"data":    user,
		},
	)
}

// Update data
func UpdateUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	type UpdateDataInput struct {
		Fullname        string `gorm:"not null" json:"fullname"`
		Email           string `gorm:"unique; not null" json:"email"`
		Telephone       string `gorm:"unique; not null" json:"telephone"` 
		Password        string `json:"password" validate:"required"`
		PasswordConfirm string `json:"password_confirm" gorm:"-"`
		Role            string `json:"role"`
		Permission      string `json:"permission"`
		Status          bool   `json:"status"` 
	}

	var updateData UpdateDataInput

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Review your input",
				"data":    nil,
			},
		)
	}

	user := new(models.User)

	db.Where("uuid = ?", uuid).First(&user)
	user.Fullname = updateData.Fullname
	user.Email = updateData.Email
	user.Telephone = updateData.Telephone 
	user.Role = updateData.Role
	user.Permission = updateData.Permission
	user.Status = updateData.Status 

	db.Save(&user)

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "user updated success",
			"data":    user,
		},
	)
}

// Delete data
func DeleteUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	db := database.DB

	var user models.User
	db.Where("uuid = ?", uuid).First(&user)
	if user.Fullname == "" {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "No user name found",
				"data":    nil,
			},
		)
	}

	db.Delete(&user)

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "user deleted success",
			"data":    nil,
		},
	)
}