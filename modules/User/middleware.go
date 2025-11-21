package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTProtect(allowedroles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "authentication required"})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "invalid token"})
		}

		claims, err := VerifyToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "invalid or expired token"})
		}

		// Safely extract role
		roleRaw, ok := claims["role"]
		roleClaim, okStr := roleRaw.(string)
		if !ok || !okStr {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "invalid role in token"})
		}

		var allowed bool
		for _, r := range allowedroles {
			if roleClaim == r || r == "*" {
				allowed = true
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"msg": "you're not allowed to access this route"})
		}
		// Safely extract user ID
		idRaw, ok := claims["user_id"]
		idFloat, okFloat := idRaw.(float64)
		if !ok || !okFloat {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "invalid user ID in token"})
		}
		userID := uint(idFloat)

		// Store in Locals for handlers
		c.Locals("user_id", userID)
		c.Locals("role", roleClaim)
		c.Locals("claims", claims)

		return c.Next()
	}
}
