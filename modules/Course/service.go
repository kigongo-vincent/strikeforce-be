package course

import (
	"fmt"
	"strconv"

	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Create(c *fiber.Ctx, db *gorm.DB) error {

	var course Course

	if err := c.BodyParser(&course); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid course details"})
	}

	var UserID = c.Locals("user_id").(uint)
	var OrgID = organization.FindById(db, UserID)
	DeptID := department.FindById(db, OrgID)

	fmt.Println("org id: " + strconv.Itoa(int(OrgID)))
	fmt.Println("Dept id: " + strconv.Itoa(int(DeptID)))
	fmt.Println("User id: " + strconv.Itoa(int(UserID)))

	if DeptID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "you can not add a course to a department you didn't create"})
	}

	course.DepartmentID = DeptID

	if err := db.Create(&course).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add course"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "course created successfully", "data": course})

}

func FindByDept(c *fiber.Ctx, db *gorm.DB) error {

	var courses []Course

	dept := c.Query("dept")

	DeptID, err := strconv.ParseUint(dept, 10, 64)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "missing department info"})
	}

	if err := db.Where("department_id = ?", DeptID).Find(&courses).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get courses"})
	}

	return c.JSON(fiber.Map{"data": courses})

}
