package supervisor

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Core"
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateSupervisorRequest struct {
	Email   string       `json:"email"`
	Name    string       `json:"name"`
	Profile user.Profile `json:"profile,omitempty"`
}

// GenerateRandomPassword generates a secure random password (same as student)
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 12 // Default to 12 characters
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Use base64 encoding for a readable password
	password := base64.URLEncoding.EncodeToString(bytes)
	// Take only the first 'length' characters
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}

// CreateForDepartment creates a supervisor user and record for a specific department
func CreateForDepartment(db *gorm.DB, departmentID uint, req CreateSupervisorRequest) (Supervisor, string, error) {
	if req.Email == "" {
		return Supervisor{}, "", fmt.Errorf("email is required")
	}
	if req.Name == "" {
		return Supervisor{}, "", fmt.Errorf("name is required")
	}

	// Verify department exists
	var dept department.Department
	if err := db.First(&dept, departmentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Supervisor{}, "", fmt.Errorf("department not found")
		}
		return Supervisor{}, "", fmt.Errorf("failed to find department: %w", err)
	}

	// Check if user already exists
	var existingUser user.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return Supervisor{}, "", fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Generate random password
	randomPassword, err := GenerateRandomPassword(12)
	if err != nil {
		return Supervisor{}, "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Create user
	newUser := user.User{
		Email:    req.Email,
		Name:     req.Name,
		Role:     "supervisor",
		Profile:  req.Profile,
		Password: user.GenerateHash(randomPassword),
	}

	if err := db.Create(&newUser).Error; err != nil {
		return Supervisor{}, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Create supervisor record
	supervisor := Supervisor{
		UserID:       newUser.ID,
		DepartmentID: departmentID,
	}

	if err := db.Create(&supervisor).Error; err != nil {
		// Rollback: delete user if supervisor creation fails
		db.Delete(&newUser)
		return Supervisor{}, "", fmt.Errorf("failed to create supervisor record: %w", err)
	}

	// Reload supervisor with relations
	db.Preload("User").Preload("Department").First(&supervisor, supervisor.ID)
	return supervisor, randomPassword, nil
}

// FindByDepartment retrieves all accepted supervisors for a department
func FindByDepartment(c *fiber.Ctx, db *gorm.DB) error {
	var supervisors []Supervisor
	dept := c.Query("dept")

	if dept == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "missing department parameter"})
	}

	DeptID, err := strconv.ParseUint(dept, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid department ID: " + err.Error()})
	}

	if err := db.Where("department_id = ?", DeptID).
		Preload("User").
		Preload("Department").
		Preload("Department.Organization").
		Find(&supervisors).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to get supervisors: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": supervisors})
}

// GetByUserID retrieves supervisor records for a user
func GetByUserID(db *gorm.DB, userID uint) ([]Supervisor, error) {
	var supervisors []Supervisor
	if err := db.Where("user_id = ?", userID).
		Preload("User").
		Preload("Department").
		Find(&supervisors).Error; err != nil {
		return nil, err
	}
	return supervisors, nil
}

// GetByID retrieves a supervisor by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var supervisor Supervisor

	if err := db.Preload("User").Preload("Department").Preload("Department.Organization").First(&supervisor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "supervisor not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get supervisor: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": supervisor})
}

// Delete deletes a supervisor record and optionally the user
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var supervisor Supervisor

	if err := db.First(&supervisor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "supervisor not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find supervisor: " + err.Error()})
	}

	// Delete supervisor record
	if err := db.Delete(&supervisor).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete supervisor: " + err.Error()})
	}

	// Optionally delete the user account (soft delete)
	// For now, we'll just delete the supervisor record
	// Uncomment below if you want to delete the user too:
	// if err := db.Delete(&user.User{}, supervisor.UserID).Error; err != nil {
	// 	return c.Status(400).JSON(fiber.Map{"msg": "failed to delete user: " + err.Error()})
	// }

	return c.JSON(fiber.Map{"msg": "supervisor deleted successfully"})
}

// Suspend suspends a supervisor (soft delete)
func Suspend(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var supervisor Supervisor

	if err := db.First(&supervisor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "supervisor not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find supervisor: " + err.Error()})
	}

	// Soft delete the supervisor record
	if err := db.Delete(&supervisor).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to suspend supervisor: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "supervisor suspended successfully"})
}

// Create creates a supervisor for a specific department
func Create(c *fiber.Ctx, db *gorm.DB) error {
	departmentIdParam := c.Params("departmentId")
	departmentId, err := strconv.ParseUint(departmentIdParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid department ID"})
	}

	// Verify department exists
	var dept department.Department
	if err := db.First(&dept, departmentId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "department not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find department"})
	}

	var req CreateSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid supervisor details: " + err.Error()})
	}

	supervisor, password, err := CreateForDepartment(db, uint(departmentId), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	// Send password email
	baseURL := core.GetFrontendURL()
	loginURL := fmt.Sprintf("%s/auth/login", baseURL)

	if err := SendPasswordEmail(req.Email, req.Name, password, loginURL); err != nil {
		fmt.Printf("Warning: Failed to send password email to %s: %v\n", req.Email, err)
	}

	return c.Status(201).JSON(fiber.Map{"msg": "supervisor created successfully", "data": supervisor})
}
