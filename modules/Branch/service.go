package branch

import (
	"fmt"
	"strconv"

	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {
	var branch Branch

	if err := c.BodyParser(&branch); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid branch details"})
	}

	if branch.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "branch name is required"})
	}

	OrgID := organization.FindById(db, c.Locals("user_id").(uint))

	if OrgID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add branch to organization"})
	}

	branch.OrganizationID = OrgID

	if err := db.Create(&branch).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add branch: " + err.Error()})
	}

	db.Preload("Organization").First(&branch, branch.ID)

	return c.Status(201).JSON(fiber.Map{
		"data": branch,
		"msg":  "Branch created successfully",
	})
}

func FindByOrg(c *fiber.Ctx, db *gorm.DB) error {
	var OrgID uint

	// Check if organizationId is provided as query parameter
	if orgIdParam := c.Query("organizationId"); orgIdParam != "" {
		orgIdUint, err := strconv.ParseUint(orgIdParam, 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid organizationId parameter"})
		}
		OrgID = uint(orgIdUint)
	} else if universityIdParam := c.Query("universityId"); universityIdParam != "" {
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

	var branches []Branch

	if err := db.Where("organization_id = ?", OrgID).Preload("Organization").Find(&branches).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get branches: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": branches})
}

func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var branch Branch

	if err := db.Preload("Organization").First(&branch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "branch not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get branch: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": branch})
}

func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var branch Branch
	if err := db.Preload("Organization").First(&branch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "branch not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find branch"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	if branch.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update this branch"})
	}

	var updateData Branch
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	if updateData.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "branch name is required"})
	}

	// Don't allow changing organization_id
	updateData.OrganizationID = branch.OrganizationID

	if err := db.Model(&branch).Updates(updateData).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update branch: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Organization").First(&branch, branch.ID)

	return c.JSON(fiber.Map{
		"msg":  "branch updated successfully",
		"data": branch,
	})
}

func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var branch Branch
	if err := db.Preload("Organization").First(&branch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "branch not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find branch"})
	}

	orgID := organization.FindById(db, userID)
	if orgID == 0 {
		return c.Status(403).JSON(fiber.Map{"msg": "unable to resolve your organization"})
	}

	if branch.OrganizationID != orgID {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to delete this branch"})
	}

	// Check if branch is being used by students
	var studentCount int64
	if err := db.Table("students").Where("branch_id = ?", id).Count(&studentCount).Error; err == nil && studentCount > 0 {
		return c.Status(400).JSON(fiber.Map{"msg": fmt.Sprintf("cannot delete branch: %d student(s) are associated with this branch", studentCount)})
	}

	if err := db.Delete(&branch).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete branch: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "branch deleted successfully"})
}

// GetStats returns statistics for branches
func GetStats(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)

	var uni organization.Organization
	if err := db.Where("user_id = ?", userID).First(&uni).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get organization"})
	}

	var totalBranches int64
	db.Model(&Branch{}).Where("organization_id = ?", uni.ID).Count(&totalBranches)

	var totalStudents int64
	db.Table("students").
		Joins("LEFT JOIN branches ON students.branch_id = branches.id").
		Where("branches.organization_id = ?", uni.ID).
		Count(&totalStudents)

	var totalProjects int64
	db.Table("projects").
		Joins("INNER JOIN students ON projects.user_id = students.user_id").
		Joins("LEFT JOIN branches ON students.branch_id = branches.id").
		Where("branches.organization_id = ?", uni.ID).
		Count(&totalProjects)

	stats := fiber.Map{
		"totalBranches": totalBranches,
		"totalStudents": totalStudents,
		"totalProjects": totalProjects,
	}

	return c.JSON(fiber.Map{"data": stats})
}

// GetStudentsByBranch returns data for students by branch graph
func GetStudentsByBranch(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)

	var uni organization.Organization
	if err := db.Where("user_id = ?", userID).First(&uni).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get organization"})
	}

	type Result struct {
		BranchName string `json:"branchName"`
		Count      int64  `json:"count"`
	}

	var results []Result
	db.Table("branches").
		Select("branches.name as branch_name, COUNT(students.id) as count").
		Joins("LEFT JOIN students ON students.branch_id = branches.id").
		Where("branches.organization_id = ?", uni.ID).
		Group("branches.id, branches.name").
		Order("count DESC").
		Scan(&results)

	// Format for frontend
	data := make([]fiber.Map, len(results))
	for i, r := range results {
		data[i] = fiber.Map{
			"branch": r.BranchName,
			"count":  r.Count,
		}
	}

	return c.JSON(fiber.Map{"data": data})
}

// GetProjectsByBranch returns data for projects by branch graph
func GetProjectsByBranch(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)

	var uni organization.Organization
	if err := db.Where("user_id = ?", userID).First(&uni).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get organization"})
	}

	type Result struct {
		BranchName string `json:"branchName"`
		Count      int64  `json:"count"`
	}

	var results []Result
	db.Table("branches").
		Select("branches.name as branch_name, COUNT(DISTINCT projects.id) as count").
		Joins("LEFT JOIN students ON students.branch_id = branches.id").
		Joins("LEFT JOIN projects ON projects.user_id = students.user_id").
		Where("branches.organization_id = ?", uni.ID).
		Group("branches.id, branches.name").
		Order("count DESC").
		Scan(&results)

	// Format for frontend
	data := make([]fiber.Map, len(results))
	for i, r := range results {
		data[i] = fiber.Map{
			"branch": r.BranchName,
			"count":  r.Count,
		}
	}

	return c.JSON(fiber.Map{"data": data})
}

