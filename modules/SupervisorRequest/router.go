package supervisorrequest

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	requests := r.Group("/supervisor-requests", user.JWTProtect([]string{"student", "supervisor", "university-admin"}))

	requests.Get("/", func(c *fiber.Ctx) error {
		return GetAll(c, db)
	})

	requests.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	requests.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	requests.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	requests.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})
}





