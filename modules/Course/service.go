package course

import (
	"fmt"
	"strconv"

	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {
	// Parse request body - frontend sends snake_case field names
	type CreateRequest struct {
		DepartmentID uint   `json:"department_id"`
		Name         string `json:"name"`
	}

	var req CreateRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid course details"})
	}

	var UserID = c.Locals("user_id").(uint)
	var OrgID = organization.FindById(db, UserID)

	if OrgID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to resolve organization"})
	}

	// Use the department_id from the request body
	if req.DepartmentID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "department_id is required"})
	}

	// Validate that the department belongs to the user's organization
	var dept department.Department
	if err := db.Where("id = ? AND organization_id = ?", req.DepartmentID, OrgID).First(&dept).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(403).JSON(fiber.Map{"msg": "you can not add a course to a department you didn't create"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to validate department: " + err.Error()})
	}

	// Build course from request
	course := Course{
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
	}

	fmt.Println("org id: " + strconv.Itoa(int(OrgID)))
	fmt.Println("Dept id: " + strconv.Itoa(int(course.DepartmentID)))
	fmt.Println("User id: " + strconv.Itoa(int(UserID)))

	if err := db.Create(&course).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add course"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "course created successfully", "data": course})

}

func FindByDept(c *fiber.Ctx, db *gorm.DB) error {

	var courses []Course

	dept := c.Query("dept")

	DeptID, err := strconv.ParseUint(dept, 10, 64)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "missing department info"})
	}

	if err := db.Where("department_id = ?", DeptID).Find(&courses).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get courses"})
	}

	return c.JSON(fiber.Map{"data": courses})
}

// GetByID retrieves a course by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var course Course

	if err := db.Preload("Department").First(&course, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "course not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get course: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": course})
}

// Update updates a course
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var course Course
	if err := db.Preload("Department.Organization").First(&course, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "course not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find course"})
	}

	// Check if user owns the organization
	if course.Department.Organization.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update this course"})
	}

	var updateData Course
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	// Don't allow changing department_id
	updateData.DepartmentID = course.DepartmentID

	if err := db.Model(&course).Updates(updateData).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update course: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Department").First(&course, course.ID)

	return c.JSON(fiber.Map{
		"msg":  "course updated successfully",
		"data": course,
	})
}

// Delete deletes a course
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var course Course
	if err := db.Preload("Department.Organization").First(&course, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "course not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find course"})
	}

	// Check if user owns the organization
	if course.Department.Organization.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to delete this course"})
	}

	if err := db.Delete(&course).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete course: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "course deleted successfully"})
}
