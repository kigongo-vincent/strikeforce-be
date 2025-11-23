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

	protected := user.Group("/", JWTProtect([]string{"*"}))
	protected.Get("/", func(c *fiber.Ctx) error {
		// id := c.Locals("user_id")
		return c.SendStatus(202)
	})

	groups := protected.Group("/group")

	groups.Post("/", func(c *fiber.Ctx) error {
		return CreateGroup(c, db)
	})

	groups.Post("/add", func(c *fiber.Ctx) error {
		return AddToGroup(c, db)
	})

	groups.Post("/remove", func(c *fiber.Ctx) error {
		return RemoveFromGroup(c, db)
	})

}
