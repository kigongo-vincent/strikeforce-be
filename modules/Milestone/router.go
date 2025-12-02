package milestone

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {

	milestones := r.Group("/milestones", user.JWTProtect([]string{"partner", "university-admin", "super-admin"}))

	milestones.Get("/", func(c *fiber.Ctx) error {
		return GetAll(c, db)
	})

	milestones.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	milestones.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	milestones.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	milestones.Put("/update-status", func(c *fiber.Ctx) error {
		return UpdateStatus(c, db)
	})

	milestones.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

}
