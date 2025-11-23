package student

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {

	students := r.Group("/students", user.JWTProtect([]string{"university-admin", "student"}))

	students.Post("/", func(c *fiber.Ctx) error {

		return Create(c, db)

	})

	students.Get("/", func(c *fiber.Ctx) error {
		return FindByCourse(c, db)
	})

}
