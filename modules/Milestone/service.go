package milestone

import (
	"strconv"

	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	var ms Milestone
	projectId := c.Query("project")

	if err := c.BodyParser(&ms); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid milestone details"})
	}

	ProjectID, _ := strconv.ParseUint(projectId, 10, 64)

	var project project.Project
	if err := db.Preload("Department.Organization.User").First(&project, ProjectID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get project" + err.Error()})
	}

	ms.ProjectID = uint(ProjectID)

	UserID := c.Locals("user_id")
	if UserID.(uint) != project.Department.Organization.UserID {
		return c.Status(400).JSON(fiber.Map{"msg": "not authorized to add milestones"})
	}

	if err := db.Create(&ms).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"data": project})

}

func UpdateStatus(c *fiber.Ctx, db *gorm.DB) error {

	status := c.Query("status")
	ms := c.Query("ms")

	MilestoneID, _ := strconv.ParseUint(ms, 10, 64)

	UserID := c.Locals("user_id").(uint)
	var milestone Milestone
	if err := db.Preload("Project.Department.Organization").First(&milestone, MilestoneID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get mielstone" + err.Error()})
	}
	if UserID != milestone.Project.Department.Organization.UserID {
		// return c.Status(400).JSON(fiber.Map{"msg": "not authorized to add milestones"})
	}

	if err := db.Model(&milestone).Update("status", status).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "" + err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "", "data": milestone})

}
