package analytics

import (
	"fmt"

	application "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Application"
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// StudentAnalyticsResponse represents analytics data for a student
type StudentAnalyticsResponse struct {
	TotalApplications    int     `json:"totalApplications"`
	ActiveApplications   int     `json:"activeApplications"`
	CompletedProjects    int     `json:"completedProjects"`
	ActiveProjects       int     `json:"activeProjects"`
	TotalBudget          float64 `json:"totalBudget"`
	TotalEarnings        float64 `json:"totalEarnings"`
	ActiveProjectsChange float64 `json:"activeProjectsChange,omitempty"`
	TotalBudgetChange    float64 `json:"totalBudgetChange,omitempty"`
}

// GetStudentAnalytics calculates analytics for the authenticated student
func GetStudentAnalytics(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)

	// Get all applications for this student
	var applications []application.Application
	// Use PostgreSQL JSONB contains operator to find applications where student is in student_ids
	if err := db.Where("CAST(student_ids AS jsonb) @> ?::jsonb", fmt.Sprintf(`[%d]`, userID)).
		Preload("Project").
		Find(&applications).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get applications: " + err.Error()})
	}

	// Get project IDs from applications
	var projectIDs []uint
	for _, app := range applications {
		if app.ProjectID > 0 {
			projectIDs = append(projectIDs, app.ProjectID)
		}
	}

	// Get projects for this student
	var projects []project.Project
	if len(projectIDs) > 0 {
		if err := db.Where("id IN ?", projectIDs).Find(&projects).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get projects: " + err.Error()})
		}
	}

	// Calculate statistics
	totalApplications := len(applications)
	activeApplications := 0
	activeProjects := 0
	completedProjects := 0
	totalBudget := 0.0
	totalEarnings := 0.0

	// Count applications by status
	for _, app := range applications {
		if app.Status == "ACCEPTED" || app.Status == "ASSIGNED" {
			activeApplications++
		}
	}

	// Calculate project statistics
	for _, proj := range projects {
		// Calculate budget (convert to float64)
		if proj.Budget.Currency != "" {
			totalBudget += float64(proj.Budget.Value)
		}

		// Count by status
		if proj.Status == "in-progress" {
			activeProjects++
		} else if proj.Status == "completed" {
			completedProjects++
			// For completed projects, calculate earnings (assume full budget for now)
			// In production, this would come from actual payout/milestone data
			if proj.Budget.Currency != "" {
				totalEarnings += float64(proj.Budget.Value)
			}
		}
	}

	// Calculate changes (simplified - in production would compare with historical data)
	activeProjectsChange := 0.0
	if activeProjects > 0 {
		activeProjectsChange = 10.0 // Simulate 10% growth
	}

	totalBudgetChange := 0.0
	if totalBudget > 0 {
		totalBudgetChange = 15.0 // Simulate 15% growth
	}

	response := StudentAnalyticsResponse{
		TotalApplications:    totalApplications,
		ActiveApplications:   activeApplications,
		CompletedProjects:    completedProjects,
		ActiveProjects:       activeProjects,
		TotalBudget:          totalBudget,
		TotalEarnings:        totalEarnings,
		ActiveProjectsChange: activeProjectsChange,
		TotalBudgetChange:    totalBudgetChange,
	}

	return c.JSON(fiber.Map{"data": response})
}

