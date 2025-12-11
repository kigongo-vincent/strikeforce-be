package college

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	// GET endpoints accessible by partners and university-admins
	colleges := r.Group("/colleges", user.JWTProtect([]string{"partner", "university-admin", "super-admin"}))

	colleges.Get("/", func(c *fiber.Ctx) error {
		return FindByOrg(c, db)
	})

	colleges.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	// POST, PUT, DELETE endpoints only for university-admins
	collegesAdmin := r.Group("/colleges", user.JWTProtect([]string{"university-admin", "delegated-admin", "super-admin"}))

	collegesAdmin.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	collegesAdmin.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	collegesAdmin.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})
}
