package organization

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UploadLogo handles logo file uploads for organizations
func UploadLogo(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)

	// Get organization for this user
	var org Organization
	if err := db.Where("user_id = ?", userID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "organization not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find organization"})
	}

	// Get uploaded file
	file, err := c.FormFile("logo")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "no logo file provided: " + err.Error()})
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	fileHeader, err := file.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to open file"})
	}
	defer fileHeader.Close()

	buffer := make([]byte, 512)
	_, err = fileHeader.Read(buffer)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to read file"})
	}

	contentType := http.DetectContentType(buffer)
	if !allowedTypes[contentType] {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid file type. Only JPEG, PNG, GIF, and WebP are allowed"})
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"msg": "file size exceeds 5MB limit"})
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/organizations/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to create upload directory"})
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("org_%d_%d%s", org.ID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to save file: " + err.Error()})
	}

	// Delete old logo if exists
	if org.Logo != "" {
		oldPath := filepath.Join("uploads", org.Logo)
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}

	// Update organization with new logo path
	org.Logo = filePath
	if err := db.Save(&org).Error; err != nil {
		// If update fails, delete the uploaded file
		os.Remove(filePath)
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update organization logo: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"msg": "logo uploaded successfully",
		"data": fiber.Map{"logo": filePath},
	})
}

