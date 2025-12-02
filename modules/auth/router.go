package auth

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers password reset endpoints under /api/auth.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	auth := app.Group("/api/auth")

	auth.Post("/forgot-password", func(c *fiber.Ctx) error {
		return ForgotPassword(c, db)
	})

	auth.Post("/reset-password", func(c *fiber.Ctx) error {
		return ResetPassword(c, db)
	})
}



