package project

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	var project Project

	if err := c.BodyParser(&project); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get project details"})
	}

	project.UserID = c.Locals("user_id").(uint)

	if err := db.Create(&project).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add project"})
	}

	data := fiber.Map{
		"msg":  "project created successfully",
		"data": project,
	}

	return c.Status(201).JSON(data)

}

func GetByOwner(c *fiber.Ctx, db *gorm.DB) error {

	var projects []Project

	if err := db.Where("user_id = ?", c.Locals("user_id").(uint)).Find(&projects).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get projects"})
	}

	return c.JSON(fiber.Map{"data": projects})

}

func Update(c *fiber.Ctx, db *gorm.DB) error {

	var project Project

	if err := c.BodyParser(&project); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid project details"})
	}

	if err := db.Updates(&project).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update"})
	}

	return c.JSON(fiber.Map{"data": project})
}

func UpdateStatus(c *fiber.Ctx, db *gorm.DB) error {

	status := c.Query("status")
	project := c.Query("project")

	ProjectID, _ := strconv.ParseUint(project, 10, 64)

	var tmp Project

	if err := db.First(&tmp, ProjectID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid project " + err.Error()})
	}
	if err := db.Model(&tmp).Update("status", status).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to update project status"})
	}
	return c.JSON(fiber.Map{"data": tmp})

}
