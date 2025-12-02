package chat

import (
	"strconv"

	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func FindAll(c *fiber.Ctx, db *gorm.DB) error {
	var messages []Message
	groupId := c.Params("group")
	if groupId == "" {
		// Try threadId if group is empty
		groupId = c.Params("threadId")
	}
	
	GroupID, err := strconv.ParseUint(groupId, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid group ID"})
	}

	if err := db.Where("group_id = ?", GroupID).Preload("Sender").Order("created_at ASC").Find(&messages).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get the messages for the provided group: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": messages})
}

func Create(c *fiber.Ctx, db *gorm.DB) error {
	var message Message
	UserID := c.Locals("user_id").(uint)

	if err := c.BodyParser(&message); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid message content"})
	}

	message.SenderID = UserID

	// Check whether the group exists
	var groupCount int64
	if err := db.Model(&user.Group{}).Where("id = ?", message.GroupID).Count(&groupCount).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to validate group"})
	}
	if groupCount == 0 {
		return c.Status(404).JSON(fiber.Map{"msg": "group not found"})
	}

	if err := db.Create(&message).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to send message: " + err.Error()})
	}

	// Load sender info
	db.Preload("Sender").First(&message, message.ID)

	// Broadcast via WebSocket if hub is initialized
	if chatHub != nil {
		chatHub.BroadcastMessage(message.GroupID, message)
	}

	return c.Status(201).JSON(fiber.Map{"data": message})
}

// GetThreadsByUser gets all chat threads (groups) for a user
func GetThreadsByUser(c *fiber.Ctx, db *gorm.DB) error {
	// Get userID from JWT token (route is protected, so token is required)
	// Never accept user ID from query parameters or request body for security
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(401).JSON(fiber.Map{"msg": "authentication required"})
	}
	
	var userID uint
	// Handle different possible types from JWT middleware
	switch v := userIDVal.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case float64:
		userID = uint(v)
	case string:
		parsed, parseErr := strconv.ParseUint(v, 10, 32)
		if parseErr != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid user_id in token: " + parseErr.Error()})
		}
		userID = uint(parsed)
	default:
		return c.Status(400).JSON(fiber.Map{"msg": "invalid user_id type in authentication token"})
	}

	// Get groups where user is leader or member
	var groups []user.Group
	if err := db.Where("user_id = ? OR id IN (SELECT group_id FROM user_groups WHERE user_id = ?)", userID, userID).
		Preload("User").Preload("Members").Find(&groups).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get groups: " + err.Error()})
	}

	// Transform to thread format
	type Thread struct {
		ID             uint   `json:"id"`
		ProjectID      uint   `json:"projectId"`
		Type           string `json:"type"`
		ParticipantIDs []uint `json:"participantIds"`
		CreatedAt      string `json:"createdAt"`
		UpdatedAt      string `json:"updatedAt"`
	}

	var threads []Thread
	for _, group := range groups {
		var participantIDs []uint
		participantIDs = append(participantIDs, group.UserID)
		for _, member := range group.Members {
			participantIDs = append(participantIDs, member.ID)
		}

		threads = append(threads, Thread{
			ID:             group.ID,
			ProjectID:      0, // Groups may not have project_id, adjust if needed
			Type:           "PROJECT",
			ParticipantIDs: participantIDs,
			CreatedAt:      group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:      group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(fiber.Map{"data": threads})
}

// GetMessagesByThread gets messages for a thread (group)
func GetMessagesByThread(c *fiber.Ctx, db *gorm.DB) error {
	var messages []Message
	threadId := c.Params("threadId")
	if threadId == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "threadId is required"})
	}
	
	ThreadID, err := strconv.ParseUint(threadId, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid thread ID"})
	}

	if err := db.Where("group_id = ?", ThreadID).Preload("Sender").Order("created_at ASC").Find(&messages).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get the messages for the provided thread: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": messages})
}
