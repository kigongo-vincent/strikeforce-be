package supervisor

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	supervisors := r.Group("/supervisors", user.JWTProtect([]string{"university-admin", "supervisor"}))

	supervisors.Get("/", func(c *fiber.Ctx) error {
		return FindByDepartment(c, db)
	})

	supervisors.Post("/:departmentId", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	supervisors.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	supervisors.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

	supervisors.Post("/:id/suspend", func(c *fiber.Ctx) error {
		return Suspend(c, db)
	})
}
