package chat

import (
	"strconv"

	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func FindAll(c *fiber.Ctx, db *gorm.DB) error {
	var messages []Message
	// Try groupId first, then group, then threadId for backward compatibility
	groupId := c.Params("groupId")
	if groupId == "" {
		groupId = c.Params("group")
	}
	if groupId == "" {
		groupId = c.Params("threadId")
	}

	if groupId == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "group ID is required"})
	}

	GroupID, err := strconv.ParseUint(groupId, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid group ID"})
	}

	if err := db.Where("group_id = ?", GroupID).Preload("Sender").Order("created_at ASC").Find(&messages).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get the messages for the provided group: " + err.Error()})
	}

	// Always return data, even if empty array
	return c.JSON(fiber.Map{"data": messages})
}

func Create(c *fiber.Ctx, db *gorm.DB) error {
	// Request struct to handle both snake_case (group_id) and camelCase (groupId) from frontend
	type CreateMessageRequest struct {
		GroupID  uint   `json:"groupId"`  // camelCase
		Group_ID uint   `json:"group_id"` // snake_case (for compatibility)
		Body     string `json:"body"`
		Type     string `json:"type,omitempty"`
	}

	var req CreateMessageRequest
	UserID := c.Locals("user_id").(uint)

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid message content"})
	}

	// Use group_id if provided, otherwise use groupId
	var groupID uint
	if req.Group_ID > 0 {
		groupID = req.Group_ID
	} else if req.GroupID > 0 {
		groupID = req.GroupID
	} else {
		return c.Status(400).JSON(fiber.Map{"msg": "groupId or group_id is required"})
	}

	// Check whether the group exists
	var group user.Group
	if err := db.First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "group not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to validate group: " + err.Error()})
	}

	// Create message
	message := Message{
		SenderID: UserID,
		GroupID:  groupID,
		Body:     req.Body,
	}

	if err := db.Create(&message).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to send message: " + err.Error()})
	}

	// Load sender info (Profile is embedded, not a relation, so no need to preload it separately)
	db.Preload("Sender").First(&message, message.ID)

	// WebSocket broadcasting removed - using plain HTTP only
	// Clients can poll or refresh to get new messages

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
		// Get project ID from application (groups are linked to projects via applications)
		var projectID uint
		var application struct {
			ProjectID uint
		}
		// Find the most recent assigned application for this group
		if err := db.Table("applications").
			Where("group_id = ? AND status = ?", group.ID, "ASSIGNED").
			Order("updated_at DESC").
			Limit(1).
			Select("project_id").
			Scan(&application).Error; err == nil && application.ProjectID > 0 {
			projectID = application.ProjectID
		} else {
			// Fallback: try to get any application for this group (not just ASSIGNED)
			if err := db.Table("applications").
				Where("group_id = ?", group.ID).
				Order("updated_at DESC").
				Limit(1).
				Select("project_id").
				Scan(&application).Error; err == nil && application.ProjectID > 0 {
				projectID = application.ProjectID
			}
		}

		var participantIDs []uint
		participantIDs = append(participantIDs, group.UserID)
		for _, member := range group.Members {
			participantIDs = append(participantIDs, member.ID)
		}

		threads = append(threads, Thread{
			ID:             group.ID,
			ProjectID:      projectID,
			Type:           "PROJECT",
			ParticipantIDs: participantIDs,
			CreatedAt:      group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:      group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Always return data, even if empty array
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

// GetMessagesByProject gets messages for a project by:
// 1. Finding the ASSIGNED application for the project
// 2. Getting the group ID from that application
// 3. Querying messages with that group ID
func GetMessagesByProject(c *fiber.Ctx, db *gorm.DB) error {
	projectId := c.Params("projectId")
	if projectId == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "project ID is required"})
	}

	ProjectID, err := strconv.ParseUint(projectId, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid project ID"})
	}

	// Step 1: Find the ASSIGNED application for this project
	var application struct {
		GroupID *uint
	}
	if err := db.Table("applications").
		Where("project_id = ? AND status = ?", ProjectID, "ASSIGNED").
		Order("updated_at DESC").
		Limit(1).
		Select("group_id").
		Scan(&application).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No assigned application found - return empty messages array
			return c.JSON(fiber.Map{"data": []Message{}})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find assigned application: " + err.Error()})
	}

	// Step 2: Check if group ID exists
	if application.GroupID == nil || *application.GroupID == 0 {
		// No group assigned yet - return empty messages array
		return c.JSON(fiber.Map{"data": []Message{}})
	}

	// Step 3: Query messages with that group ID
	var messages []Message
	if err := db.Where("group_id = ?", *application.GroupID).
		Preload("Sender").
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get messages for the group: " + err.Error()})
	}

	// Always return data, even if empty array
	return c.JSON(fiber.Map{"data": messages})
}
