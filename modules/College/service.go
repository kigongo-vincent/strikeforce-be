package college

import (
	"fmt"
	"strconv"

	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Create a college for the current organization (university admin)
func Create(c *fiber.Ctx, db *gorm.DB) error {
	var college College

	if err := c.BodyParser(&college); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid college details"})
	}

	if college.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "college name is required"})
	}

	orgID := organization.FindById(db, c.Locals("user_id").(uint))
	if orgID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add college to organization"})
	}

	college.OrganizationID = orgID

	if err := db.Create(&college).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add college: " + err.Error()})
	}

	db.Preload("Organization").First(&college, college.ID)

	return c.Status(201).JSON(fiber.Map{
		"data": college,
		"msg":  "College created successfully",
	})
}

// FindByOrg fetches colleges for a university, supporting explicit organizationId/universityId filters for partners.
func FindByOrg(c *fiber.Ctx, db *gorm.DB) error {
	var orgID uint

	if orgIdParam := c.Query("organizationId"); orgIdParam != "" {
		orgIdUint, err := strconv.ParseUint(orgIdParam, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid organizationId parameter"})
		}
		orgID = uint(orgIdUint)
	} else if universityIdParam := c.Query("universityId"); universityIdParam != "" {
		universityIdUint, err := strconv.ParseUint(universityIdParam, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid universityId parameter"})
		}
		orgID = uint(universityIdUint)
	} else {
		var uni organization.Organization
		if err := db.Where("user_id = ?", c.Locals("user_id").(uint)).First(&uni).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get organization. Please provide organizationId or universityId query parameter"})
		}
		orgID = uni.ID
	}

	var colleges []College
	if err := db.Where("organization_id = ?", orgID).Preload("Organization").Find(&colleges).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get colleges: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": colleges})
}

// GetByID fetches a college by id
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var college College

	if err := db.Preload("Organization").First(&college, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "college not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get college: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": college})
}

// Update edits a college that belongs to the current organization
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var college College
	if err := db.Preload("Organization").First(&college, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "college not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find college"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	if college.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update this college"})
	}

	var updateData College
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	if updateData.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "college name is required"})
	}

	updateData.OrganizationID = college.OrganizationID

	if err := db.Model(&college).Updates(updateData).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update college: " + err.Error()})
	}

	db.Preload("Organization").First(&college, college.ID)

	return c.JSON(fiber.Map{
		"msg":  "college updated successfully",
		"data": college,
	})
}

// Delete removes a college when it belongs to the current organization and has no attached departments
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var college College
	if err := db.Preload("Organization").First(&college, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "college not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find college"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	if college.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to delete this college"})
	}

	// Prevent deletion if departments are attached to this college
	var deptCount int64
	if err := db.Table("departments").Where("college_id = ?", id).Count(&deptCount).Error; err == nil && deptCount > 0 {
		return c.Status(400).JSON(fiber.Map{"msg": fmt.Sprintf("cannot delete college: %d department(s) are associated with this college", deptCount)})
	}

	if err := db.Delete(&college).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete college: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "college deleted successfully"})
}

