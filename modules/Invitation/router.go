package invitation

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers invitation routes
func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	invitations := r.Group("/invitations", user.JWTProtect([]string{"university-admin", "delegated-admin", "super-admin"}))

	invitations.Get("/", func(c *fiber.Ctx) error {
		return GetAll(c, db)
	})

	invitations.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	invitations.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	invitations.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	invitations.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

	// Public endpoints (no auth required)
	public := r.Group("/invitations")
	public.Get("/token/:token", func(c *fiber.Ctx) error {
		return GetByToken(c, db)
	})

	public.Post("/accept", func(c *fiber.Ctx) error {
		return Accept(c, db)
	})
}
