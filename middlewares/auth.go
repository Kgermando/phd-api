package middlewares

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kgermando/phd-api/utils"
)

func extractToken(c *fiber.Ctx) string {
	if token := c.Query("token"); token != "" {
		return token
	}

	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}

	return ""
}

func IsAuthenticated(c *fiber.Ctx) error {

	token := extractToken(c)
	if token == "" {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	userUUID, err := utils.VerifyJwt(token)
	if err != nil {
		message := "unauthenticated"
		if errors.Is(err, jwt.ErrTokenExpired) {
			message = "token expired"
		}
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": message,
		})
	}

	// Store the user UUID in the context
	c.Locals("user_uuid", userUUID)

	c.Next()
	return nil
}
