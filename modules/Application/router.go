package application

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers application routes
func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	applications := r.Group("/applications", user.JWTProtect([]string{"*"}))

	applications.Get("/", func(c *fiber.Ctx) error {
		return GetAll(c, db)
	})

	applications.Get("/:id", func(c *fiber.Ctx) error {
		return GetByID(c, db)
	})

	applications.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})

	applications.Put("/:id", func(c *fiber.Ctx) error {
		return Update(c, db)
	})

	applications.Delete("/:id", func(c *fiber.Ctx) error {
		return Delete(c, db)
	})

	applications.Post("/upload", func(c *fiber.Ctx) error {
		return UploadFiles(c, db)
	})

	// Screening endpoints (university admin actions)
	applications.Post("/:id/score", func(c *fiber.Ctx) error {
		return ScoreApplication(c, db)
	})

	applications.Post("/:id/shortlist", func(c *fiber.Ctx) error {
		return ShortlistApplication(c, db)
	})

	applications.Post("/:id/reject", func(c *fiber.Ctx) error {
		return RejectApplication(c, db)
	})

	applications.Post("/:id/waitlist", func(c *fiber.Ctx) error {
		return WaitlistApplication(c, db)
	})

	// Offer endpoints
	applications.Post("/:id/offer", func(c *fiber.Ctx) error {
		return OfferApplication(c, db)
	})

	applications.Post("/:id/accept-offer", func(c *fiber.Ctx) error {
		return AcceptOffer(c, db)
	})

	applications.Post("/:id/decline-offer", func(c *fiber.Ctx) error {
		return DeclineOffer(c, db)
	})
}
