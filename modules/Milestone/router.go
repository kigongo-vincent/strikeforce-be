package milestone

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {

	milestones := r.Group("/milestones", user.JWTProtect([]string{"partner"}))

	milestones.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})
	milestones.Put("/update-status", func(c *fiber.Ctx) error {
		return UpdateStatus(c, db)
	})

}
