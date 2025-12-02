package invitation

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetAll retrieves all invitations with optional filters
func GetAll(c *fiber.Ctx, db *gorm.DB) error {
	var invitations []Invitation
	query := db.Model(&Invitation{})

	// Filter by universityId (organizationId)
	if universityId := c.Query("universityId"); universityId != "" {
		universityIdUint, err := strconv.ParseUint(universityId, 10, 32)
		if err == nil {
			query = query.Where("organization_id = ?", uint(universityIdUint))
		}
	}

	// Filter by departmentId
	if departmentId := c.Query("departmentId"); departmentId != "" {
		departmentIdUint, err := strconv.ParseUint(departmentId, 10, 32)
		if err == nil {
			query = query.Where("department_id = ?", uint(departmentIdUint))
		}
	}

	// Filter by role
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Preload("Organization").Preload("User").Find(&invitations).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get invitations: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": invitations})
}

// GetByID retrieves an invitation by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var invitation Invitation

	if err := db.Preload("Organization").Preload("User").First(&invitation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "invitation not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get invitation: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": invitation})
}

// GetByToken retrieves an invitation by token (public endpoint)
func GetByToken(c *fiber.Ctx, db *gorm.DB) error {
	token := c.Params("token")
	var invitation Invitation

	if err := db.Preload("Organization").Where("token = ?", token).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "invitation not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get invitation: " + err.Error()})
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = "EXPIRED"
		db.Model(&invitation).Update("status", "EXPIRED")
		return c.Status(400).JSON(fiber.Map{"msg": "invitation has expired"})
	}

	return c.JSON(fiber.Map{"data": invitation})
}

// Create creates a new invitation
func Create(c *fiber.Ctx, db *gorm.DB) error {
	var invitation Invitation
	userID := c.Locals("user_id").(uint)

	if err := c.BodyParser(&invitation); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid invitation data: " + err.Error()})
	}

	// Validate required fields
	if invitation.Email == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "email is required"})
	}
	if invitation.Role == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "role is required"})
	}
	if invitation.OrganizationID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "organization_id is required"})
	}

	// Validate role
	if invitation.Role != "student" && invitation.Role != "supervisor" {
		return c.Status(400).JSON(fiber.Map{"msg": "role must be 'student' or 'supervisor'"})
	}

	// Validate departmentId for supervisors
	if invitation.Role == "supervisor" && invitation.DepartmentID == nil {
		return c.Status(400).JSON(fiber.Map{"msg": "departmentId is required for supervisor invitations"})
	}

	// Check if user has permission to create invitations for this organization
	// (This is a simplified check - you may want to add more validation)
	var orgCount int64
	db.Model(&invitation).Where("organization_id = ? AND user_id = ?", invitation.OrganizationID, userID).Count(&orgCount)
	// For now, allow if user is university-admin or super-admin

	// Generate secure token
	token, err := GenerateToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to generate token"})
	}
	invitation.Token = token

	// Set default expiry (7 days from now)
	if invitation.ExpiresAt.IsZero() {
		invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	// Set default status
	if invitation.Status == "" {
		invitation.Status = "PENDING"
	}

	// Check if invitation already exists for this email and organization
	var existing Invitation
	if err := db.Where("email = ? AND organization_id = ? AND status = ?", invitation.Email, invitation.OrganizationID, "PENDING").First(&existing).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"msg": "pending invitation already exists for this email"})
	}

	if err := db.Create(&invitation).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create invitation: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Organization").First(&invitation, invitation.ID)

	// Send invitation email
	orgName := invitation.Organization.Name
	if orgName == "" {
		orgName = "StrikeForce"
	}
	// Use provided name or extract from email as fallback
	name := invitation.Name
	if name == "" {
		name = strings.Split(invitation.Email, "@")[0]
	}
	if err := SendInvitationEmail(invitation.Email, invitation.Token, name, invitation.Role, orgName); err != nil {
		// Log error but don't fail the request - invitation is created
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	return c.Status(201).JSON(fiber.Map{
		"msg":  "invitation created successfully",
		"data": invitation,
	})
}

// Accept accepts an invitation and creates a user account
func Accept(c *fiber.Ctx, db *gorm.DB) error {
	type AcceptRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	var req AcceptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid request data: " + err.Error()})
	}

	if req.Token == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "token is required"})
	}
	if req.Password == "" || len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"msg": "password is required and must be at least 8 characters"})
	}

	// Find invitation
	var invitation Invitation
	if err := db.Preload("Organization").Where("token = ?", req.Token).First(&invitation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "invitation not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find invitation"})
	}

	// Check if already used
	if invitation.Status == "USED" {
		return c.Status(400).JSON(fiber.Map{"msg": "invitation has already been used"})
	}

	// Check if expired
	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = "EXPIRED"
		db.Model(&invitation).Update("status", "EXPIRED")
		return c.Status(400).JSON(fiber.Map{"msg": "invitation has expired"})
	}

	// Check if user already exists with this email
	var existingUser user.User
	if err := db.Where("email = ?", invitation.Email).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"msg": "user with this email already exists"})
	}

	// Validate name - use invitation name if provided, otherwise require it in request
	nameToUse := strings.TrimSpace(req.Name)
	if nameToUse == "" {
		// Fallback to name stored in invitation if available
		nameToUse = strings.TrimSpace(invitation.Name)
	}
	if nameToUse == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "name is required"})
	}

	// Create user
	newUser := user.User{
		Email:    invitation.Email,
		Role:     invitation.Role,
		Name:     nameToUse,
		Password: user.GenerateHash(req.Password),
	}

	if err := db.Create(&newUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create user: " + err.Error()})
	}

	// Reload user to ensure all fields are populated (including profile)
	if err := db.First(&newUser, newUser.ID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to reload user: " + err.Error()})
	}

	// Note: Supervisors are now created directly, not via invitations
	// This invitation acceptance is only for students

	// Mark invitation as used
	now := time.Now()
	invitation.Status = "USED"
	invitation.UsedAt = &now
	invitation.UserID = &newUser.ID
	if err := db.Save(&invitation).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update invitation: " + err.Error()})
	}

	// Generate token for new user
	token, err := user.GenerateToken(newUser)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to generate token"})
	}

	newUser.Password = "" // Don't return password

	return c.Status(201).JSON(fiber.Map{
		"msg": "invitation accepted and user created successfully",
		"data": fiber.Map{
			"user":  newUser,
			"token": token,
		},
	})
}

// Update updates an existing invitation
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var invitation Invitation

	if err := db.First(&invitation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "invitation not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find invitation"})
	}

	var updateData Invitation
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	// Don't allow updating token or used invitations
	if invitation.Status == "USED" {
		return c.Status(400).JSON(fiber.Map{"msg": "cannot update used invitation"})
	}

	// Update fields (exclude ID, token, and timestamps)
	updateData.Token = invitation.Token // Preserve token
	if err := db.Model(&invitation).Updates(updateData).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update invitation: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Organization").Preload("User").First(&invitation, invitation.ID)

	return c.JSON(fiber.Map{
		"msg":  "invitation updated successfully",
		"data": invitation,
	})
}

// Delete deletes an invitation
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var invitation Invitation

	if err := db.First(&invitation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "invitation not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find invitation"})
	}

	if err := db.Delete(&invitation).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete invitation: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "invitation deleted successfully"})
}
