package branch

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	// GET endpoints accessible by partners and university-admins
	branches := r.Group("/branches", user.JWTProtect([]string{"partner", "university-admin", "delegated-admin", "super-admin"}))

	branches.Get("/", func(c *fiber.Ctx) error {
		return FindByOrg(c, db)
	})

	branches.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	branches.Get("/stats/summary", func(c *fiber.Ctx) error {
		return GetStats(c, db)
	})

	branches.Get("/stats/students-by-branch", func(c *fiber.Ctx) error {
		return GetStudentsByBranch(c, db)
	})

	branches.Get("/stats/projects-by-branch", func(c *fiber.Ctx) error {
		return GetProjectsByBranch(c, db)
	})

	// POST, PUT, DELETE endpoints only for university-admins
	branchesAdmin := r.Group("/branches", user.JWTProtect([]string{"university-admin", "delegated-admin", "super-admin"}))

	branchesAdmin.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	branchesAdmin.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	branchesAdmin.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})
}
