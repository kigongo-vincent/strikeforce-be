package user

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {

	user := app.Group("/user")

	user.Post("/login", func(c *fiber.Ctx) error {
		return Login(c, db)
	})
	user.Post("/signup", func(c *fiber.Ctx) error {
		return SignUp(c, db)
	})

	user.Get("/verify", Verify)

	protected := user.Group("/", JWTProtect([]string{"company_admin", "university_admin"}))
	protected.Get("/", func(c *fiber.Ctx) error {
		// id := c.Locals("user_id")
		return c.SendStatus(202)
	})
}
