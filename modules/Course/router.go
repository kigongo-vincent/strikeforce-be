package course

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {

	courses := r.Group("/courses", user.JWTProtect([]string{"university-admin"}))

	courses.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	courses.Get("/", func(c *fiber.Ctx) error {
		return FindByDept(c, db)
	})

}
