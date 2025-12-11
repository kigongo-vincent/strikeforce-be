package chat

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var chatHub *Hub

// InitHub initializes the chat hub (call this from main.go)
func InitHub() {
	chatHub = NewHub()
	go chatHub.Run()
}

func RegisterRoutes(r fiber.Router, db *gorm.DB) {
	if chatHub == nil {
		InitHub()
	}

	chats := r.Group("/chats", user.JWTProtect([]string{"*"}))

	// Specific routes must come before parameterized routes
	// WebSocket endpoint
	chats.Get("/ws", func(c *fiber.Ctx) error {
		return HandleWebSocket(c, db, chatHub)
	})

	// Get chat threads for user (must come before /:group)
	chats.Get("/threads", func(c *fiber.Ctx) error {
		return GetThreadsByUser(c, db)
	})

	// Get messages for a thread (must come before /:group)
	chats.Get("/threads/:threadId/messages", func(c *fiber.Ctx) error {
		return GetMessagesByThread(c, db)
	})

	// Get messages by project ID (finds ASSIGNED application, gets group, then messages)
	chats.Get("/project/:projectId", func(c *fiber.Ctx) error {
		return GetMessagesByProject(c, db)
	})

	// Parameterized routes come last (for backward compatibility)
	// This handles both /chats/:group and works as /chats/group/:groupId via params
	chats.Get("/:group", func(c *fiber.Ctx) error {
		return FindAll(c, db)
	})

	chats.Post("/", func(c *fiber.Ctx) error {
		return Create(c, db)
	})
}
