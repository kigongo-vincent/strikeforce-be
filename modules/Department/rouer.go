package department

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRRoutes(r fiber.Router, db *gorm.DB) {
	// GET endpoints accessible by partners and university-admins
	departments := r.Group("/departments", user.JWTProtect([]string{"partner", "university-admin", "super-admin"}))

	departments.Get("/", func(c *fiber.Ctx) error {
		return FindByOrg(c, db)
	})

	departments.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	// POST, PUT, DELETE endpoints only for university-admins
	departmentsAdmin := r.Group("/departments", user.JWTProtect([]string{"university-admin", "delegated-admin", "super-admin"}))

	departmentsAdmin.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	departmentsAdmin.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	departmentsAdmin.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

}
