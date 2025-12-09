package department

import (
	"strconv"

	college "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/College"
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	var department Department

	if err := c.BodyParser(&department); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid department details"})
	}

	OrgID := organization.FindById(db, c.Locals("user_id").(uint))

	if OrgID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add department to organization"})
	}

	if department.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "department name is required"})
	}

	// Validate college if provided (required for new departments)
	if department.CollegeID == nil || *department.CollegeID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "collegeId is required to create a department"})
	}

	var deptCollege college.College
	if err := db.First(&deptCollege, department.CollegeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "college not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find college: " + err.Error()})
	}

	if deptCollege.OrganizationID != OrgID {
		return c.Status(403).JSON(fiber.Map{"msg": "college does not belong to your organization"})
	}

	department.OrganizationID = OrgID
	department.Organization.ID = OrgID
	department.Organization.User.ID = c.Locals("user_id").(uint)
	department.Organization.UserID = c.Locals("user_id").(uint)

	if err := db.Create(&department).Error; err != nil {
		{
			return c.Status(400).JSON(fiber.Map{"msg": "failed to add department" + err.Error()})

		}
	}

	db.Preload("Organization").Preload("College").First(&department, department.ID)

	data := fiber.Map{
		"data": department,
		"msg":  "Department created successfully",
	}

	return c.Status(201).JSON(data)

}

func FindByOrg(c *fiber.Ctx, db *gorm.DB) error {
	var OrgID uint

	// Check if organizationId is provided as query parameter (for partners)
	if orgIdParam := c.Query("organizationId"); orgIdParam != "" {
		orgIdUint, err := strconv.ParseUint(orgIdParam, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid organizationId parameter"})
		}
		OrgID = uint(orgIdUint)
	} else if universityIdParam := c.Query("universityId"); universityIdParam != "" {
		// Also support universityId as an alias for organizationId
		universityIdUint, err := strconv.ParseUint(universityIdParam, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid universityId parameter"})
		}
		OrgID = uint(universityIdUint)
	} else {
		// For university-admin, get organization from logged-in user
		var uni organization.Organization
		if err := db.Where("user_id = ?", c.Locals("user_id").(uint)).First(&uni).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get organization. Please provide organizationId or universityId query parameter"})
		}
		OrgID = uni.ID
	}

	var depts []Department

	if err := db.Where("organization_id = ?", OrgID).Preload("Organization").Preload("College").Find(&depts).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get departments: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": depts})
}

// GetByID retrieves a department by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var dept Department

	if err := db.Preload("Organization").Preload("College").First(&dept, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "department not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get department: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": dept})
}

// Update updates a department
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var dept Department
	if err := db.Preload("Organization").Preload("College").First(&dept, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "department not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find department"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	// Ensure the department belongs to the same organization as the user
	if dept.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update this department"})
	}

	var updateData Department
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	if updateData.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "department name is required"})
	}

	// Validate new college if provided
	var collegeIDToUpdate *uint
	if updateData.CollegeID != nil && *updateData.CollegeID != 0 {
		var newCollege college.College
		if err := db.First(&newCollege, *updateData.CollegeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(404).JSON(fiber.Map{"msg": "college not found"})
			}
			return c.Status(400).JSON(fiber.Map{"msg": "failed to find college: " + err.Error()})
		}

		if newCollege.OrganizationID != orgID {
			return c.Status(403).JSON(fiber.Map{"msg": "college does not belong to your organization"})
		}
		collegeIDToUpdate = updateData.CollegeID
	} else {
		// Keep existing college if not provided
		collegeIDToUpdate = dept.CollegeID
	}

	// Update name and college_id explicitly
	updateFields := map[string]interface{}{
		"name": updateData.Name,
	}
	if collegeIDToUpdate != nil {
		updateFields["college_id"] = *collegeIDToUpdate
	} else {
		updateFields["college_id"] = nil
	}

	if err := db.Model(&dept).Updates(updateFields).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update department: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Organization").Preload("College").First(&dept, dept.ID)

	return c.JSON(fiber.Map{
		"msg":  "department updated successfully",
		"data": dept,
	})
}

// Delete deletes a department
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var dept Department
	if err := db.Preload("Organization").Preload("College").First(&dept, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "department not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find department"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	// Ensure the department belongs to the same organization as the user
	if dept.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to delete this department"})
	}

	if err := db.Delete(&dept).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete department: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "department deleted successfully"})
}
