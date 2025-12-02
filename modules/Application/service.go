package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GetAll retrieves all applications with optional filters
func GetAll(c *fiber.Ctx, db *gorm.DB) error {
	var applications []Application
	query := db.Model(&Application{})
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	// Filter by projectId
	if projectId := c.Query("projectId"); projectId != "" {
		projectIdUint, err := strconv.ParseUint(projectId, 10, 32)
		if err == nil {
			query = query.Where("project_id = ?", uint(projectIdUint))
		}
	}

	// Role-based filtering:
	// - Students: Only see their own applications (user ID in student_ids)
	// - Partners: See all applications for their projects
	// - University admins: See all applications for projects in their university
	// - Super-admins: See all applications
	if role == "student" {
		// Students can only see their own applications
		// Use PostgreSQL JSONB contains operator (@>) to check if array contains userID
		query = query.Where("CAST(student_ids AS jsonb) @> ?::jsonb", fmt.Sprintf(`[%d]`, userID))
	} else if role == "partner" {
		// Partners can see all applications for their projects
		// Join with projects table and filter by user_id (project owner)
		query = query.Joins("JOIN projects ON applications.project_id = projects.id").
			Where("projects.user_id = ?", userID)
	} else if role == "university-admin" {
		// University admins can see all applications for projects in their university
		// Get user's organization ID
		var userOrgID uint
		if err := db.Table("organizations").
			Where("user_id = ?", userID).
			Select("id").
			Scan(&userOrgID).Error; err != nil {
			// If no organization found, return empty results
			return c.JSON(fiber.Map{"data": []Application{}})
		}
		// Join with projects -> departments -> organizations to filter by university
		query = query.Joins("JOIN projects ON applications.project_id = projects.id").
			Joins("JOIN departments ON projects.department_id = departments.id").
			Where("departments.organization_id = ?", userOrgID)
	} else if role == "super-admin" {
		// Super-admins can see all applications (no additional filter)
	} else {
		// For other roles or unknown roles, default to student behavior (own applications only)
		query = query.Where("CAST(student_ids AS jsonb) @> ?::jsonb", fmt.Sprintf(`[%d]`, userID))
	}

	if err := query.Preload("Project").Preload("Group").Find(&applications).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get applications: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": applications})
}

// GetByID retrieves an application by ID
func GetByID(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application

	if err := db.Preload("Project").Preload("Group").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get application: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": application})
}

// Create creates a new application
func Create(c *fiber.Ctx, db *gorm.DB) error {
	var application Application
	userID := c.Locals("user_id").(uint)

	if err := c.BodyParser(&application); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid application data: " + err.Error()})
	}

	// Validate required fields
	if application.ProjectID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "project_id is required"})
	}

	// Check if project exists
	var proj project.Project
	if err := db.First(&proj, application.ProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "project not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to validate project"})
	}

	// If applicant type is INDIVIDUAL, ensure student_ids contains the current user
	if application.ApplicantType == "INDIVIDUAL" || application.ApplicantType == "" {
		application.ApplicantType = "INDIVIDUAL"
		studentIDs := []uint{userID}
		studentIDsJSON, _ := json.Marshal(studentIDs)
		application.StudentIDs = datatypes.JSON(studentIDsJSON)

		// Create a default group for individual applications
		// This ensures the application has a group from the start for chat and other group-related features
		var student user.User
		if err := db.First(&student, userID).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "student not found"})
		}

		// Create a group for the individual application
		groupName := fmt.Sprintf("%s - %s", student.Name, proj.Title)
		if len(groupName) > 100 {
			groupName = groupName[:100] // Truncate if too long
		}

		group := user.Group{
			UserID:   userID, // Student is the leader
			Name:     groupName,
			Capacity: 1,
		}

		if err := db.Create(&group).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to create default group for individual application: " + err.Error()})
		}

		// Add the student as a member
		var members []user.User
		members = append(members, student)

		if len(members) > 0 {
			if err := db.Model(&group).Association("Members").Append(members); err != nil {
				return c.Status(400).JSON(fiber.Map{"msg": "failed to add member to group: " + err.Error()})
			}
		}

		// Link the group to the application
		application.GroupID = &group.ID
	} else if application.ApplicantType == "GROUP" {
		// Validate group exists and user is a member
		if application.GroupID == nil {
			return c.Status(400).JSON(fiber.Map{"msg": "group_id is required for GROUP applications"})
		}
		var group user.Group
		if err := db.Preload("Members").First(&group, application.GroupID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"msg": "group not found"})
		}
		// Extract member IDs
		var memberIDs []uint
		for _, member := range group.Members {
			memberIDs = append(memberIDs, member.ID)
		}
		studentIDsJSON, _ := json.Marshal(memberIDs)
		application.StudentIDs = datatypes.JSON(studentIDsJSON)
	}

	// Set default status
	if application.Status == "" {
		application.Status = "SUBMITTED"
	}

	if err := db.Create(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create application: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Project").Preload("Group").Preload("Group.Members").First(&application, application.ID)

	return c.Status(201).JSON(fiber.Map{
		"msg":  "application created successfully",
		"data": application,
	})
}

// Update updates an existing application
func Update(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Authorization check for status updates and scoring
	// Only university admins, partners (project owners), and super-admins can update application status/score
	canUpdate := false
	if role == "super-admin" {
		canUpdate = true
	} else if role == "partner" {
		// Partners can only update applications for their own projects
		if application.Project.UserID == userID {
			canUpdate = true
		}
	} else if role == "university-admin" {
		// University admins can update applications for projects in their university
		// Get user's organization ID
		var userOrgID uint
		if err := db.Table("organizations").
			Where("user_id = ?", userID).
			Select("id").
			Scan(&userOrgID).Error; err == nil {
			// Get project's department to check organization
			var deptOrgID uint
			if err := db.Table("departments").
				Where("id = ?", application.Project.DepartmentID).
				Select("organization_id").
				Scan(&deptOrgID).Error; err == nil {
				if deptOrgID == userOrgID {
					canUpdate = true
				}
			}
		}
	} else if role == "student" {
		// Students can only update their own applications (for withdrawal, etc.)
		// Check if user ID is in student_ids
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err == nil {
			for _, sid := range studentIDs {
				if sid == userID {
					canUpdate = true
					break
				}
			}
		}
	}

	// Parse update data - handle both direct Application struct and partial updates
	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid update data: " + err.Error()})
	}

	// Check if trying to update status or score without permission
	if (updateData["status"] != nil || updateData["score"] != nil) && !canUpdate {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update application status or score"})
	}

	// Validate and update status if provided
	if statusVal, ok := updateData["status"].(string); ok {
		validStatuses := map[string]bool{
			"SUBMITTED": true, "SHORTLISTED": true, "WAITLIST": true,
			"REJECTED": true, "OFFERED": true, "ACCEPTED": true,
			"DECLINED": true, "ASSIGNED": true,
		}
		if !validStatuses[statusVal] {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid status value"})
		}
		application.Status = statusVal
	}

	// Update score if provided (handle as JSON)
	if scoreVal, ok := updateData["score"]; ok {
		scoreJSON, err := json.Marshal(scoreVal)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid score data: " + err.Error()})
		}
		application.Score = datatypes.JSON(scoreJSON)
	}

	// Update other fields if provided
	if statementVal, ok := updateData["statement"].(string); ok {
		application.Statement = statementVal
	}
	if attachmentsVal, ok := updateData["attachments"]; ok {
		attachmentsJSON, err := json.Marshal(attachmentsVal)
		if err == nil {
			application.Attachments = datatypes.JSON(attachmentsJSON)
		}
	}
	// Handle offerExpiresAt update
	if offerExpiresAtVal, ok := updateData["offerExpiresAt"].(string); ok && offerExpiresAtVal != "" {
		expiryDate, err := time.Parse("2006-01-02T15:04:05.000Z", offerExpiresAtVal)
		if err != nil {
			// Try alternative format
			expiryDate, err = time.Parse("2006-01-02", offerExpiresAtVal)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"msg": "invalid offerExpiresAt format"})
			}
		}
		expiryDateVal := datatypes.Date(expiryDate)
		application.OfferExpiresAt = &expiryDateVal
	}

	// Update updatedAt timestamp
	application.UpdatedAt = time.Now()

	// Save updates
	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update application: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Project").Preload("Group").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  "application updated successfully",
		"data": application,
	})
}

// ScoreApplication scores an application (university admin action)
func ScoreApplication(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Authorization: Only university admins, partners, and super-admins can score
	canScore := false
	if role == "super-admin" {
		canScore = true
	} else if role == "partner" {
		if application.Project.UserID == userID {
			canScore = true
		}
	} else if role == "university-admin" {
		var userOrgID uint
		if err := db.Table("organizations").
			Where("user_id = ?", userID).
			Select("id").
			Scan(&userOrgID).Error; err == nil {
			var deptOrgID uint
			if err := db.Table("departments").
				Where("id = ?", application.Project.DepartmentID).
				Select("organization_id").
				Scan(&deptOrgID).Error; err == nil {
				if deptOrgID == userOrgID {
					canScore = true
				}
			}
		}
	}

	if !canScore {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to score this application"})
	}

	type ScoreRequest struct {
		Score map[string]interface{} `json:"score"`
	}

	var req ScoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid score data: " + err.Error()})
	}

	// Marshal score to JSON
	scoreJSON, err := json.Marshal(req.Score)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid score format: " + err.Error()})
	}

	application.Score = datatypes.JSON(scoreJSON)
	application.UpdatedAt = time.Now()

	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update application score: " + err.Error()})
	}

	db.Preload("Project").Preload("Group").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  "application scored successfully",
		"data": application,
	})
}

// ShortlistApplication shortlists an application (university admin action)
func ShortlistApplication(c *fiber.Ctx, db *gorm.DB) error {
	return updateApplicationStatus(c, db, "SHORTLISTED")
}

// RejectApplication rejects an application (university admin action)
func RejectApplication(c *fiber.Ctx, db *gorm.DB) error {
	return updateApplicationStatus(c, db, "REJECTED")
}

// WaitlistApplication waitlists an application (university admin action)
func WaitlistApplication(c *fiber.Ctx, db *gorm.DB) error {
	return updateApplicationStatus(c, db, "WAITLIST")
}

// updateApplicationStatus is a helper function to update application status with authorization
func updateApplicationStatus(c *fiber.Ctx, db *gorm.DB, newStatus string) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Authorization check
	canUpdate := false
	if role == "super-admin" {
		canUpdate = true
	} else if role == "partner" {
		if application.Project.UserID == userID {
			canUpdate = true
		}
	} else if role == "university-admin" {
		var userOrgID uint
		if err := db.Table("organizations").
			Where("user_id = ?", userID).
			Select("id").
			Scan(&userOrgID).Error; err == nil {
			var deptOrgID uint
			if err := db.Table("departments").
				Where("id = ?", application.Project.DepartmentID).
				Select("organization_id").
				Scan(&deptOrgID).Error; err == nil {
				if deptOrgID == userOrgID {
					canUpdate = true
				}
			}
		}
	}

	if !canUpdate {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to update this application status"})
	}

	// Validate status
	validStatuses := map[string]bool{
		"SUBMITTED": true, "SHORTLISTED": true, "WAITLIST": true,
		"REJECTED": true, "OFFERED": true, "ACCEPTED": true,
		"DECLINED": true, "ASSIGNED": true,
	}
	if !validStatuses[newStatus] {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid status value"})
	}

	// If setting status to ASSIGNED, ensure application has a group for chat functionality
	if newStatus == "ASSIGNED" && application.GroupID == nil {
		// Extract student IDs from the application
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to parse student IDs"})
		}

		if len(studentIDs) == 0 {
			return c.Status(400).JSON(fiber.Map{"msg": "application must have at least one student"})
		}

		// Get the first student (leader)
		var student user.User
		if err := db.First(&student, studentIDs[0]).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "student not found"})
		}

		// Create a group for the application
		groupName := fmt.Sprintf("%s - %s", student.Name, application.Project.Title)
		if len(groupName) > 100 {
			groupName = groupName[:100] // Truncate if too long
		}

		group := user.Group{
			UserID:   studentIDs[0], // First student is the leader
			Name:     groupName,
			Capacity: len(studentIDs),
		}

		if err := db.Create(&group).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to create group for application: " + err.Error()})
		}

		// Add all students as members
		var members []user.User
		for _, studentID := range studentIDs {
			var member user.User
			if err := db.First(&member, studentID).Error; err == nil {
				members = append(members, member)
			}
		}

		if len(members) > 0 {
			if err := db.Model(&group).Association("Members").Append(members); err != nil {
				return c.Status(400).JSON(fiber.Map{"msg": "failed to add members to group: " + err.Error()})
			}
		}

		// Link the group to the application
		application.GroupID = &group.ID
	} else if newStatus == "ASSIGNED" && application.GroupID != nil {
		// Verify the group still exists
		var groupCount int64
		if err := db.Model(&user.Group{}).Where("id = ?", application.GroupID).Count(&groupCount).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to validate group"})
		}
		if groupCount == 0 {
			return c.Status(404).JSON(fiber.Map{"msg": "group not found. Please reassign the application."})
		}
	}

	application.Status = newStatus
	application.UpdatedAt = time.Now()

	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update application status: " + err.Error()})
	}

	db.Preload("Project").Preload("Group").Preload("Group.Members").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  fmt.Sprintf("application %s successfully", newStatus),
		"data": application,
	})
}

// OfferApplication issues an offer to an application (university admin action)
func OfferApplication(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Authorization: Only university admins, partners, and super-admins can offer
	canOffer := false
	if role == "super-admin" {
		canOffer = true
	} else if role == "partner" {
		if application.Project.UserID == userID {
			canOffer = true
		}
	} else if role == "university-admin" {
		var userOrgID uint
		if err := db.Table("organizations").
			Where("user_id = ?", userID).
			Select("id").
			Scan(&userOrgID).Error; err == nil {
			var deptOrgID uint
			if err := db.Table("departments").
				Where("id = ?", application.Project.DepartmentID).
				Select("organization_id").
				Scan(&deptOrgID).Error; err == nil {
				if deptOrgID == userOrgID {
					canOffer = true
				}
			}
		}
	}

	if !canOffer {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have permission to offer this application"})
	}

	// Validate that application is in SHORTLISTED status (only shortlisted applications can receive offers)
	if application.Status != "SHORTLISTED" {
		return c.Status(400).JSON(fiber.Map{"msg": fmt.Sprintf("cannot offer application with status %s. Application must be SHORTLISTED to receive an offer", application.Status)})
	}

	// Check if project already has an assigned application
	var existingAssignedApplication Application
	if err := db.Where("project_id = ? AND status = ?", application.ProjectID, "ASSIGNED").First(&existingAssignedApplication).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"msg": "this project already has an assigned group"})
	}

	// Ensure all assigned applications have a group for chat functionality
	// Chat requires a group, so we create one if it doesn't exist
	if application.GroupID == nil {
		// Extract student IDs from the application
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to parse student IDs"})
		}

		if len(studentIDs) == 0 {
			return c.Status(400).JSON(fiber.Map{"msg": "application must have at least one student"})
		}

		// Get the first student (leader for individual, or first member for group)
		var student user.User
		if err := db.First(&student, studentIDs[0]).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "student not found"})
		}

		// Create a group for the application
		// Group name will be based on the student's name and project
		var groupName string
		if application.ApplicantType == "GROUP" {
			// For group applications, try to use existing group name if available
			groupName = fmt.Sprintf("%s - %s", student.Name, application.Project.Title)
		} else {
			// For individual applications, use student name and project
			groupName = fmt.Sprintf("%s - %s", student.Name, application.Project.Title)
		}

		if len(groupName) > 100 {
			groupName = groupName[:100] // Truncate if too long
		}

		group := user.Group{
			UserID:   studentIDs[0], // First student is the leader
			Name:     groupName,
			Capacity: len(studentIDs),
		}

		if err := db.Create(&group).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to create group for application: " + err.Error()})
		}

		// Add all students as members
		var members []user.User
		for _, studentID := range studentIDs {
			var member user.User
			if err := db.First(&member, studentID).Error; err == nil {
				members = append(members, member)
			}
		}

		if len(members) > 0 {
			if err := db.Model(&group).Association("Members").Append(members); err != nil {
				return c.Status(400).JSON(fiber.Map{"msg": "failed to add members to group: " + err.Error()})
			}
		}

		// Link the group to the application
		application.GroupID = &group.ID
	} else {
		// For GROUP applications, verify the group still exists
		var groupCount int64
		if err := db.Model(&user.Group{}).Where("id = ?", application.GroupID).Count(&groupCount).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to validate group"})
		}
		if groupCount == 0 {
			return c.Status(404).JSON(fiber.Map{"msg": "group not found. Please reassign the application."})
		}
	}

	// Update application status to ASSIGNED (immediate assignment, no offer/accept flow)
	application.Status = "ASSIGNED"
	application.UpdatedAt = time.Now()

	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to assign application: " + err.Error()})
	}

	// Reload with relations
	db.Preload("Project").Preload("Group").Preload("Group.Members").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  "group assigned to project successfully",
		"data": application,
	})
}

// AcceptOffer accepts an offer (student action)
func AcceptOffer(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Only students can accept offers (or partners/university admins on their behalf)
	if role == "student" {
		// Verify the student is part of this application
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid application data"})
		}

		isApplicant := false
		for _, sid := range studentIDs {
			if sid == userID {
				isApplicant = true
				break
			}
		}

		if !isApplicant {
			return c.Status(403).JSON(fiber.Map{"msg": "you can only accept offers for your own applications"})
		}
	} else if role != "super-admin" && role != "university-admin" && role != "partner" {
		return c.Status(403).JSON(fiber.Map{"msg": "only students can accept offers"})
	}

	// Validate that application has an active offer
	if application.Status != "OFFERED" {
		return c.Status(400).JSON(fiber.Map{"msg": "application does not have an active offer"})
	}

	// Check if offer has expired
	if application.OfferExpiresAt != nil {
		expiryTime := time.Time(*application.OfferExpiresAt)
		if time.Now().After(expiryTime) {
			return c.Status(400).JSON(fiber.Map{"msg": "this offer has expired"})
		}
	}

	// Ensure application has a group for chat functionality before assigning
	if application.GroupID == nil {
		// Extract student IDs from the application
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to parse student IDs"})
		}

		if len(studentIDs) == 0 {
			return c.Status(400).JSON(fiber.Map{"msg": "application must have at least one student"})
		}

		// Get the first student (leader)
		var student user.User
		if err := db.First(&student, studentIDs[0]).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "student not found"})
		}

		// Create a group for the application
		groupName := fmt.Sprintf("%s - %s", student.Name, application.Project.Title)
		if len(groupName) > 100 {
			groupName = groupName[:100] // Truncate if too long
		}

		group := user.Group{
			UserID:   studentIDs[0], // First student is the leader
			Name:     groupName,
			Capacity: len(studentIDs),
		}

		if err := db.Create(&group).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to create group for application: " + err.Error()})
		}

		// Add all students as members
		var members []user.User
		for _, studentID := range studentIDs {
			var member user.User
			if err := db.First(&member, studentID).Error; err == nil {
				members = append(members, member)
			}
		}

		if len(members) > 0 {
			if err := db.Model(&group).Association("Members").Append(members); err != nil {
				return c.Status(400).JSON(fiber.Map{"msg": "failed to add members to group: " + err.Error()})
			}
		}

		// Link the group to the application
		application.GroupID = &group.ID
	} else {
		// Verify the group still exists
		var groupCount int64
		if err := db.Model(&user.Group{}).Where("id = ?", application.GroupID).Count(&groupCount).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to validate group"})
		}
		if groupCount == 0 {
			return c.Status(404).JSON(fiber.Map{"msg": "group not found. Please contact support."})
		}
	}

	// Update status to ASSIGNED (student accepted the offer)
	application.Status = "ASSIGNED"
	application.UpdatedAt = time.Now()

	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to accept offer: " + err.Error()})
	}

	db.Preload("Project").Preload("Group").Preload("Group.Members").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  "offer accepted successfully",
		"data": application,
	})
}

// DeclineOffer declines an offer (student action)
func DeclineOffer(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	if err := db.Preload("Project").First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	// Only students can decline offers
	if role == "student" {
		// Verify the student is part of this application
		var studentIDs []uint
		if err := json.Unmarshal(application.StudentIDs, &studentIDs); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "invalid application data"})
		}

		isApplicant := false
		for _, sid := range studentIDs {
			if sid == userID {
				isApplicant = true
				break
			}
		}

		if !isApplicant {
			return c.Status(403).JSON(fiber.Map{"msg": "you can only decline offers for your own applications"})
		}
	} else if role != "super-admin" {
		return c.Status(403).JSON(fiber.Map{"msg": "only students can decline offers"})
	}

	// Validate that application has an offer
	if application.Status != "OFFERED" {
		return c.Status(400).JSON(fiber.Map{"msg": "application does not have an active offer"})
	}

	// Update status to DECLINED (student declined the offer)
	application.Status = "DECLINED"
	application.UpdatedAt = time.Now()

	if err := db.Save(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to decline offer: " + err.Error()})
	}

	db.Preload("Project").Preload("Group").First(&application, application.ID)

	return c.JSON(fiber.Map{
		"msg":  "offer declined successfully",
		"data": application,
	})
}

// Delete deletes an application
func Delete(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var application Application

	if err := db.First(&application, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "application not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find application"})
	}

	if err := db.Delete(&application).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to delete application: " + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "application deleted successfully"})
}

// UploadFiles handles file uploads for applications
func UploadFiles(c *fiber.Ctx, db *gorm.DB) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to parse form: " + err.Error()})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "no files provided"})
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/applications"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to create upload directory"})
	}

	var paths []string
	for _, file := range files {
		// Generate unique filename
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		filePath := filepath.Join(uploadDir, filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to save file: " + err.Error()})
		}

		paths = append(paths, filePath)
	}

	return c.JSON(fiber.Map{
		"msg":  "files uploaded successfully",
		"data": fiber.Map{"paths": paths},
	})
}
