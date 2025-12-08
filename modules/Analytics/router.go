package analytics

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	analytics := r.Group("/analytics", user.JWTProtect([]string{"student"}))

	// Student analytics endpoint
	analytics.Get("/student", func(c *fiber.Ctx) error {
		return GetStudentAnalytics(c, db)
	})

	// University admin analytics endpoint
	universityAnalytics := r.Group("/analytics/university", user.JWTProtect([]string{"university-admin"}))
	universityAnalytics.Get("/:id", func(c *fiber.Ctx) error {
		return GetUniversityAdminAnalytics(c, db)
	})
}





