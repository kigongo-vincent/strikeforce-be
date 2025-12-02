package portfolio

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GetAll retrieves all portfolio items with optional filters
func GetAll(c *fiber.Ctx, db *gorm.DB) error {
	var items []PortfolioItem
	query := db.Model(&PortfolioItem{})

	// Filter by userId (from query param or JWT token)
	if userId := c.Query("userId"); userId != "" {
		userIdUint, err := strconv.ParseUint(userId, 10, 32)
		if err == nil {
			query = query.Where("user_id = ?", uint(userIdUint))
		}
	} else {
		// If no userId in query, use authenticated user's ID
		userID := c.Locals("user_id").(uint)
		query = query.Where("user_id = ?", userID)
	}

	// Filter by projectId
	if projectId := c.Query("projectId"); projectId != "" {
		projectIdUint, err := strconv.ParseUint(projectId, 10, 32)
		if err == nil {
			query = query.Where("project_id = ?", uint(projectIdUint))
		}
	}

	if err := query.Preload("User").Preload("Project").Order("verified_at DESC").Find(&items).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get portfolio items: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": items})
}

// GetByID retrieves a portfolio item by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var item PortfolioItem

	if err := db.Preload("User").Preload("Project").First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "portfolio item not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get portfolio item: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": item})
}

// Create creates a new portfolio item
func Create(c *fiber.Ctx, db *gorm.DB) error {
	type CreateRequest struct {
		ProjectID       uint     `json:"projectId"`
		MilestoneID     *uint    `json:"milestoneId,omitempty"`
		Role            string   `json:"role"`
		Scope           string   `json:"scope"`
		Proof           []string `json:"proof"`
		Rating          *float64 `json:"rating,omitempty"`
		Complexity      string   `json:"complexity"` // LOW, MEDIUM, HIGH
		AmountDelivered float64  `json:"amountDelivered"`
		Currency        string   `json:"currency"`
		OnTime          bool     `json:"onTime"`
	}

	var req CreateRequest
	userID := c.Locals("user_id").(uint)

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid request data: " + err.Error()})
	}

	// Validate required fields
	if req.ProjectID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "project_id is required"})
	}

	if req.Role == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "role is required"})
	}

	if req.Scope == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "scope is required"})
	}

	// Validate complexity
	validComplexities := map[string]bool{"LOW": true, "MEDIUM": true, "HIGH": true}
	if !validComplexities[req.Complexity] {
		return c.Status(400).JSON(fiber.Map{"msg": "complexity must be LOW, MEDIUM, or HIGH"})
	}

	// Convert proof array to JSON
	var proofJSON []byte
	if len(req.Proof) > 0 {
		proofJSON, _ = json.Marshal(req.Proof)
	} else {
		proofJSON, _ = json.Marshal([]string{})
	}

	// Create portfolio item
	item := PortfolioItem{
		UserID:          userID,
		ProjectID:       req.ProjectID,
		MilestoneID:     req.MilestoneID,
		Role:            req.Role,
		Scope:           req.Scope,
		Proof:           datatypes.JSON(proofJSON),
		Rating:          req.Rating,
		Complexity:      req.Complexity,
		AmountDelivered: req.AmountDelivered,
		Currency:        req.Currency,
		OnTime:          req.OnTime,
		VerifiedAt:      datatypes.Date(time.Now()),
	}

	if err := db.Create(&item).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create portfolio item: " + err.Error()})
	}

	// Preload relations for response
	if err := db.Preload("User").Preload("Project").First(&item, item.ID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to load portfolio item details: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"data": item})
}

// Update updates a portfolio item
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var item PortfolioItem

	if err := db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "portfolio item not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get portfolio item: " + err.Error()})
	}

	type UpdateRequest struct {
		Role            *string   `json:"role"`
		Scope           *string   `json:"scope"`
		Proof           []string  `json:"proof"`
		Rating          *float64  `json:"rating"`
		Complexity      *string   `json:"complexity"`
		AmountDelivered *float64  `json:"amountDelivered"`
		Currency        *string   `json:"currency"`
		OnTime          *bool     `json:"onTime"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid request data: " + err.Error()})
	}

	// Update fields
	if req.Role != nil {
		item.Role = *req.Role
	}
	if req.Scope != nil {
		item.Scope = *req.Scope
	}
	if req.Proof != nil {
		proofJSON, _ := json.Marshal(req.Proof)
		item.Proof = datatypes.JSON(proofJSON)
	}
	if req.Rating != nil {
		item.Rating = req.Rating
	}
	if req.Complexity != nil {
		validComplexities := map[string]bool{"LOW": true, "MEDIUM": true, "HIGH": true}
		if !validComplexities[*req.Complexity] {
			return c.Status(400).JSON(fiber.Map{"msg": "complexity must be LOW, MEDIUM, or HIGH"})
		}
		item.Complexity = *req.Complexity
	}
	if req.AmountDelivered != nil {
		item.AmountDelivered = *req.AmountDelivered
	}
	if req.Currency != nil {
		item.Currency = *req.Currency
	}
	if req.OnTime != nil {
		item.OnTime = *req.OnTime
	}

	if err := db.Save(&item).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update portfolio item: " + err.Error()})
	}

	// Preload relations for response
	if err := db.Preload("User").Preload("Project").First(&item, item.ID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to load portfolio item details: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": item})
}

// Delete deletes a portfolio item
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var item PortfolioItem

	if err := db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "portfolio item not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get portfolio item: " + err.Error()})
	}

	if err := db.Delete(&item).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete portfolio item: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "portfolio item deleted successfully"})
}

