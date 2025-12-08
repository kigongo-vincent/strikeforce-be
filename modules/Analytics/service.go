package analytics

import (
	"fmt"
	"strconv"
	"time"

	application "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Application"
	portfolio "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Portfolio"
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	student "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Student"
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

// UniversityAdminAnalyticsResponse represents analytics data for a university admin
type UniversityAdminAnalyticsResponse struct {
	// Revenue metrics
	TotalRevenue          float64                    `json:"totalRevenue"`
	RevenueTrend          []RevenueTrendPoint        `json:"revenueTrend"`
	
	// Student earnings metrics
	TotalStudentEarnings  float64                    `json:"totalStudentEarnings"`
	EarningsTrend         []EarningsTrendPoint       `json:"earningsTrend"`
	StudentsEarningCount  int                        `json:"studentsEarningCount"`
	
	// Demographic analytics
	GenderDistribution    []GenderDistributionPoint  `json:"genderDistribution"`
	BranchDistribution    []BranchDistributionPoint  `json:"branchDistribution"`
	DistrictDistribution  []DistrictDistributionPoint `json:"districtDistribution"`
}

type RevenueTrendPoint struct {
	Month string  `json:"month"`
	Value float64 `json:"value"`
}

type EarningsTrendPoint struct {
	Month string  `json:"month"`
	Value float64 `json:"value"`
}

type GenderDistributionPoint struct {
	Gender string `json:"gender"`
	Count  int    `json:"count"`
}

type BranchDistributionPoint struct {
	BranchName string `json:"branchName"`
	Count      int    `json:"count"`
}

type DistrictDistributionPoint struct {
	District string `json:"district"`
	Count    int    `json:"count"`
}

// GetUniversityAdminAnalytics calculates analytics for a university
func GetUniversityAdminAnalytics(c *fiber.Ctx, db *gorm.DB) error {
	idParam := c.Params("id")
	orgID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "invalid organization id"})
	}

	// Get all students for this university
	var students []student.Student
	if err := db.Table("students").
		Joins("JOIN courses ON students.course_id = courses.id").
		Joins("JOIN departments ON courses.department_id = departments.id").
		Where("departments.organization_id = ?", orgID).
		Preload("User").
		Preload("Branch").
		Find(&students).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get students: " + err.Error()})
	}

	// Get student user IDs
	var studentUserIDs []uint
	for _, s := range students {
		studentUserIDs = append(studentUserIDs, s.UserID)
	}

	// Get all portfolio items for these students (for earnings calculation)
	var portfolioItems []portfolio.PortfolioItem
	if len(studentUserIDs) > 0 {
		if err := db.Where("user_id IN ?", studentUserIDs).
			Preload("Project").
			Find(&portfolioItems).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get portfolio items: " + err.Error()})
		}
	}

	// Get all projects for this university (for revenue calculation)
	var projects []project.Project
	if err := db.Table("projects").
		Joins("JOIN departments ON projects.department_id = departments.id").
		Where("departments.organization_id = ?", orgID).
		Find(&projects).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get projects: " + err.Error()})
	}

	// Calculate total revenue from completed projects
	totalRevenue := 0.0
	for _, proj := range projects {
		if proj.Status == "completed" && proj.Budget.Currency != "" {
			totalRevenue += float64(proj.Budget.Value)
		}
	}

	// Calculate total student earnings from portfolio items
	totalStudentEarnings := 0.0
	studentsWithEarnings := make(map[uint]bool)
	for _, item := range portfolioItems {
		totalStudentEarnings += item.AmountDelivered
		studentsWithEarnings[item.UserID] = true
	}

	// Calculate revenue trend (last 12 months)
	revenueTrend := calculateRevenueTrend(projects)

	// Calculate earnings trend (last 12 months)
	earningsTrend := calculateEarningsTrend(portfolioItems)

	// Calculate gender distribution
	genderDist := calculateGenderDistribution(students)

	// Calculate branch distribution
	branchDist := calculateBranchDistribution(students)

	// Calculate district distribution
	districtDist := calculateDistrictDistribution(students)

	response := UniversityAdminAnalyticsResponse{
		TotalRevenue:         totalRevenue,
		RevenueTrend:         revenueTrend,
		TotalStudentEarnings: totalStudentEarnings,
		EarningsTrend:        earningsTrend,
		StudentsEarningCount: len(studentsWithEarnings),
		GenderDistribution:   genderDist,
		BranchDistribution:   branchDist,
		DistrictDistribution: districtDist,
	}

	return c.JSON(fiber.Map{"data": response})
}

// calculateRevenueTrend calculates revenue over the last 12 months
func calculateRevenueTrend(projects []project.Project) []RevenueTrendPoint {
	now := time.Now()
	trend := make([]RevenueTrendPoint, 12)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for i := 0; i < 12; i++ {
		targetMonth := now.AddDate(0, -11+i, 0)
		monthStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		revenue := 0.0
		for _, proj := range projects {
			if proj.Status == "completed" && proj.Budget.Currency != "" {
				completedAt := proj.UpdatedAt // Use UpdatedAt as completion time
				if completedAt.After(monthStart) && completedAt.Before(monthEnd) {
					revenue += float64(proj.Budget.Value)
				}
			}
		}

		trend[i] = RevenueTrendPoint{
			Month: fmt.Sprintf("%s %d", monthNames[targetMonth.Month()-1], targetMonth.Year()),
			Value: revenue,
		}
	}

	return trend
}

// calculateEarningsTrend calculates student earnings over the last 12 months
func calculateEarningsTrend(portfolioItems []portfolio.PortfolioItem) []EarningsTrendPoint {
	now := time.Now()
	trend := make([]EarningsTrendPoint, 12)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for i := 0; i < 12; i++ {
		targetMonth := now.AddDate(0, -11+i, 0)
		monthStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		earnings := 0.0
		for _, item := range portfolioItems {
			verifiedAt := time.Time(item.VerifiedAt)
			// Skip if VerifiedAt is zero (not set)
			if verifiedAt.IsZero() {
				continue
			}
			if verifiedAt.After(monthStart) && verifiedAt.Before(monthEnd) {
				earnings += item.AmountDelivered
			}
		}

		trend[i] = EarningsTrendPoint{
			Month: fmt.Sprintf("%s %d", monthNames[targetMonth.Month()-1], targetMonth.Year()),
			Value: earnings,
		}
	}

	return trend
}

// calculateGenderDistribution calculates gender distribution
func calculateGenderDistribution(students []student.Student) []GenderDistributionPoint {
	genderMap := make(map[string]int)
	for _, s := range students {
		gender := s.Gender
		if gender == "" {
			gender = "Not Specified"
		}
		genderMap[gender]++
	}

	var dist []GenderDistributionPoint
	for gender, count := range genderMap {
		dist = append(dist, GenderDistributionPoint{
			Gender: gender,
			Count:  count,
		})
	}

	return dist
}

// calculateBranchDistribution calculates branch distribution
func calculateBranchDistribution(students []student.Student) []BranchDistributionPoint {
	branchMap := make(map[string]int)
	for _, s := range students {
		branchName := "Not Specified"
		if s.Branch != nil && s.Branch.Name != "" {
			branchName = s.Branch.Name
		} else if s.UniversityBranch != "" {
			branchName = s.UniversityBranch
		}
		branchMap[branchName]++
	}

	var dist []BranchDistributionPoint
	for branchName, count := range branchMap {
		dist = append(dist, BranchDistributionPoint{
			BranchName: branchName,
			Count:      count,
		})
	}

	return dist
}

// calculateDistrictDistribution calculates district distribution
func calculateDistrictDistribution(students []student.Student) []DistrictDistributionPoint {
	districtMap := make(map[string]int)
	for _, s := range students {
		district := s.District
		if district == "" {
			district = "Not Specified"
		}
		districtMap[district]++
	}

	var dist []DistrictDistributionPoint
	for district, count := range districtMap {
		dist = append(dist, DistrictDistributionPoint{
			District: district,
			Count:    count,
		})
	}

	return dist
}

