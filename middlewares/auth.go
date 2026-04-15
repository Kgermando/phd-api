package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/dentic-support-api/utils"
)

func IsAuthenticated(c *fiber.Ctx) error {

	token := c.Query("token")

	fmt.Println("Token:", token)

	userUUID, err := utils.VerifyJwt(token)
	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	// Store the user UUID in the context
	c.Locals("user_uuid", userUUID)

	c.Next()
	return nil
}
