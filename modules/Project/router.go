package project

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	projects := r.Group("/projects", user.JWTProtect([]string{"partner"}))
	projects.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})
	projects.Get("/mine", func(c *fiber.Ctx) error {
		return GetByOwner(c, db)
	})
	projects.Put("/update", func(c *fiber.Ctx) error {
		return Update(c, db)
	})
	projects.Put("/update-status", func(c *fiber.Ctx) error {
		return UpdateStatus(c, db)
	})
}
