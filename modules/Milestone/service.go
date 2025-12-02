package milestone

import (
	"strconv"

	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {
	// Parse request body - frontend sends projectId in body or query
	type CreateRequest struct {
		ProjectID          *uint   `json:"projectId,omitempty"`
		Title              string  `json:"title"`
		Scope              string  `json:"scope"`
		AcceptanceCriteria string  `json:"acceptanceCriteria,omitempty"`
		DueDate            string  `json:"dueDate"`
		Amount             int     `json:"amount"`
		Currency           string  `json:"currency,omitempty"`
		Status             string  `json:"status,omitempty"`
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid milestone details: " + err.Error()})
	}

	// Get projectId from body, query parameter, or error
	var projectIdStr string
	if req.ProjectID != nil {
		projectIdStr = strconv.FormatUint(uint64(*req.ProjectID), 10)
	} else {
		projectIdStr = c.Query("project")
		if projectIdStr == "" {
			projectIdStr = c.Query("projectId")
		}
	}

	if projectIdStr == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "projectId is required"})
	}

	ProjectID, err := strconv.ParseUint(projectIdStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid projectId"})
	}

	var project project.Project
	if err := db.Preload("Department").Preload("Department.Organization").Preload("User").First(&project, ProjectID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get project: " + err.Error()})
	}

	// Check authorization - allow project owner (UserID) or super-admin
	UserID := c.Locals("user_id").(uint)
	userRole := c.Locals("role")
	
	// Project owner (partner who created the project) can add milestones
	// Super-admin can also add milestones
	isAuthorized := UserID == project.UserID || userRole == "super-admin"
	
	if !isAuthorized {
		return c.Status(403).JSON(fiber.Map{"msg": "not authorized to add milestones. Only project owner can add milestones."})
	}

	// Build milestone from request
	ms := Milestone{
		ProjectID:          uint(ProjectID),
		Title:              req.Title,
		Scope:              req.Scope,
		AcceptanceCriteria: req.AcceptanceCriteria,
		DueDate:            req.DueDate,
		Amount:             req.Amount,
		Currency:           req.Currency,
		Status:             req.Status,
	}

	// Set defaults if not provided
	if ms.AcceptanceCriteria == "" {
		ms.AcceptanceCriteria = "To be defined"
	}
	if ms.Status == "" {
		ms.Status = "PROPOSED"
	}
	if ms.Currency == "" {
		// Use project currency if available
		ms.Currency = "UGX" // Default fallback
	}

	if err := db.Create(&ms).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create milestone: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Project").First(&ms, ms.ID)

	return c.Status(201).JSON(fiber.Map{"msg": "milestone created successfully", "data": ms})
}

func UpdateStatus(c *fiber.Ctx, db *gorm.DB) error {

	status := c.Query("status")
	ms := c.Query("ms")

	MilestoneID, _ := strconv.ParseUint(ms, 10, 64)

	UserID := c.Locals("user_id").(uint)
	userRole := c.Locals("role")
	var milestone Milestone
	if err := db.Preload("Project").First(&milestone, MilestoneID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get milestone: " + err.Error()})
	}
	// Check authorization - allow project owner or super-admin
	isAuthorized := UserID == milestone.Project.UserID || userRole == "super-admin"
	if !isAuthorized {
		return c.Status(403).JSON(fiber.Map{"msg": "not authorized to update milestone status"})
	}

	if err := db.Model(&milestone).Update("status", status).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "" + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "milestone status updated successfully", "data": milestone})
}

// GetAll retrieves all milestones with optional filters
func GetAll(c *fiber.Ctx, db *gorm.DB) error {
	var milestones []Milestone
	query := db.Model(&Milestone{})

	// Filter by projectId
	if projectId := c.Query("projectId"); projectId != "" {
		projectIdUint, err := strconv.ParseUint(projectId, 10, 32)
		if err == nil {
			query = query.Where("project_id = ?", uint(projectIdUint))
		}
	}

	if err := query.Preload("Project").Find(&milestones).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get milestones: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": milestones})
}

// GetByID retrieves a milestone by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var milestone Milestone

	if err := db.Preload("Project").First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "milestone not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get milestone: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": milestone})
}

// Update updates a milestone
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var milestone Milestone

	if err := db.Preload("Project").First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "milestone not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find milestone"})
	}

	// Check authorization - allow project owner or super-admin
	UserID := c.Locals("user_id").(uint)
	userRole := c.Locals("role")
	isAuthorized := UserID == milestone.Project.UserID || userRole == "super-admin"
	if !isAuthorized {
		return c.Status(403).JSON(fiber.Map{"msg": "not authorized to update this milestone"})
	}

	// Parse update request with camelCase fields
	type UpdateRequest struct {
		Title              string `json:"title,omitempty"`
		Scope              string `json:"scope,omitempty"`
		AcceptanceCriteria string `json:"acceptanceCriteria,omitempty"`
		DueDate            string `json:"dueDate,omitempty"`
		Amount             *int   `json:"amount,omitempty"`
		Currency           string `json:"currency,omitempty"`
		Status             string `json:"status,omitempty"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	// Update only provided fields
	if req.Title != "" {
		milestone.Title = req.Title
	}
	if req.Scope != "" {
		milestone.Scope = req.Scope
	}
	if req.AcceptanceCriteria != "" {
		milestone.AcceptanceCriteria = req.AcceptanceCriteria
	}
	if req.DueDate != "" {
		milestone.DueDate = req.DueDate
	}
	if req.Amount != nil {
		milestone.Amount = *req.Amount
	}
	if req.Currency != "" {
		milestone.Currency = req.Currency
	}
	if req.Status != "" {
		milestone.Status = req.Status
	}

	// Don't allow changing project_id
	// Save the original ProjectID before update
	originalProjectID := milestone.ProjectID

	if err := db.Model(&milestone).Updates(milestone).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update milestone: " + err.Error()})
	}

	// Ensure ProjectID wasn't changed
	if milestone.ProjectID != originalProjectID {
		db.Model(&milestone).Update("project_id", originalProjectID)
	}

	// Reload with relations
	db.Preload("Project").First(&milestone, milestone.ID)

	return c.JSON(fiber.Map{
		"msg":  "milestone updated successfully",
		"data": milestone,
	})
}

// Delete deletes a milestone
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var milestone Milestone

	if err := db.Preload("Project").First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "milestone not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find milestone"})
	}

	// Check authorization - allow project owner or super-admin
	UserID := c.Locals("user_id").(uint)
	userRole := c.Locals("role")
	isAuthorized := UserID == milestone.Project.UserID || userRole == "super-admin"
	if !isAuthorized {
		return c.Status(403).JSON(fiber.Map{"msg": "not authorized to delete this milestone"})
	}

	if err := db.Delete(&milestone).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete milestone: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "milestone deleted successfully"})
}
