package project

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	projects := r.Group("/projects", user.JWTProtect([]string{"partner", "student", "university-admin", "super-admin"}))

	projects.Get("/", func(c *fiber.Ctx) error {
		return GetAll(c, db)
	})

	projects.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	projects.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	projects.Put("/update", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	projects.Put("/update-status", func(c *fiber.Ctx) error {
		return UpdateStatus(c, db)
	})

	projects.Put("/assign-supervisor", func(c *fiber.Ctx) error {
		return AssignSupervisor(c, db)
	})

	projects.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

	projects.Post("/upload", func(c *fiber.Ctx) error {
		return UploadFiles(c, db)
	})

	// Keep existing endpoint for backward compatibility
	projects.Get("/mine", func(c *fiber.Ctx) error {
		return GetByOwner(c, db)
	})
}
