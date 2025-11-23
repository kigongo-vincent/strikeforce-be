package student

import (
	"strconv"

	course "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Course"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	var student Student
	if err := c.BodyParser(&student); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid student details"})
	}

	UserID := c.Locals("user_id").(uint)
	CreatorID := course.FindById(db, student.CourseID)

	if UserID != CreatorID {
		return c.Status(403).JSON(fiber.Map{"msg": "you're not allowed to add students"})
	}

	tmpUser := student.User
	tmpUser.Role = "student"

	if tmpUser.Email == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "provide the email"})
	}

	if tmpUser.Name == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "provide your full name"})
	}

	if err := db.Where("email = ?", tmpUser.Email).First(&tmpUser).Error; err == nil {
		return c.Status(402).JSON(fiber.Map{"msg": "user with email " + tmpUser.Email + " already exists"})
	}

	student.UserID = tmpUser.ID

	if err := db.Create(&student).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add student"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "student created successfully", "data": student})
}

func FindByCourse(c *fiber.Ctx, db *gorm.DB) error {

	var students []Student
	course := c.Query("course")

	CourseID, err := strconv.ParseUint(course, 10, 64)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "missing course info"})
	}

	if err := db.Where("course_id = ?", CourseID).Find(&students).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get students"})
	}

	return c.JSON(fiber.Map{"data": students})

}
