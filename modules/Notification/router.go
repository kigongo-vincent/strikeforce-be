package notification

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {

	notifications := r.Group("/notifications", user.JWTProtect([]string{"*"}))

	notifications.Get("/", func(c *fiber.Ctx) error {
		return FindAll(c, db)
	})

	notifications.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	notifications.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	notifications.Put("/:notification", func(c *fiber.Ctx) error {
		return MarkSeen(c, db)
	})

	notifications.Patch("/mark-all-read", func(c *fiber.Ctx) error {
		return MarkAllAsRead(c, db)
	})

}
