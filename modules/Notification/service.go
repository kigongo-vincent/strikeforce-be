package notification

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	UserID := c.Locals("user_id").(uint)
	var notification Notification

	if err := c.BodyParser(&notification); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid notification details"})
	}

	notification.UserID = UserID

	if err := db.Create(&notification).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to send notification"})
	}
	return c.Status(201).JSON(fiber.Map{"data": notification})

}

func FindAll(c *fiber.Ctx, db *gorm.DB) error {

	var notifications []Notification
	UserID := c.Locals("user_id")

	if err := db.Where("user_id = ?", UserID).Find(&notifications).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get notifications"})
	}

	return c.JSON(fiber.Map{"data": notifications})

}

func MarkSeen(c *fiber.Ctx, db *gorm.DB) error {

	notificationId := c.Params("notification")
	NotificationID, _ := strconv.ParseUint(notificationId, 10, 64)

	var notification Notification
	if err := db.First(&notification, NotificationID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "notification does not exist"})
	}

	if err := db.Model(&notification).Update("seen", true).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to mark as seen"})
	}

	return c.Status(201).JSON(fiber.Map{"data": notification})
}

// GetByID retrieves a notification by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var notification Notification
	if err := db.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "notification not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get notification: " + err.Error()})
	}

	// Users can only view their own notifications
	if notification.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to view this notification"})
	}

	return c.JSON(fiber.Map{"data": notification})
}

// MarkAllAsRead marks all notifications as read for a user
func MarkAllAsRead(c *fiber.Ctx, db *gorm.DB) error {
	// Never accept userId from request body - always use authenticated user from JWT token
	userID := c.Locals("user_id").(uint)

	if err := db.Model(&Notification{}).Where("user_id = ?", userID).Update("seen", true).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to mark notifications as read: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "all notifications marked as read"})
}
