package student

import (
	"fmt"
	"strconv"

	course "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Course"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateStudentRequest struct {
	Email   string       `json:"email"`
	Name    string       `json:"name"`
	Profile user.Profile `json:"profile,omitempty"`
}

type BulkStudentsRequest struct {
	Students []CreateStudentRequest `json:"students"`
}

func createStudentForCourse(db *gorm.DB, courseId uint64, req CreateStudentRequest) (Student, string, error) {
	if req.Email == "" {
		return Student{}, "", fmt.Errorf("email is required")
	}
	if req.Name == "" {
		return Student{}, "", fmt.Errorf("name is required")
	}

	var existingUser user.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return Student{}, "", fmt.Errorf("user with email %s already exists", req.Email)
	}

	randomPassword, err := GenerateRandomPassword(12)
	if err != nil {
		return Student{}, "", fmt.Errorf("failed to generate password: %w", err)
	}

	newUser := user.User{
		Email:    req.Email,
		Name:     req.Name,
		Role:     "student",
		Profile:  req.Profile,
		Password: user.GenerateHash(randomPassword),
	}

	if err := db.Create(&newUser).Error; err != nil {
		return Student{}, "", fmt.Errorf("failed to create user: %w", err)
	}

	student := Student{
		UserID:   newUser.ID,
		CourseID: uint(courseId),
	}

	if err := db.Create(&student).Error; err != nil {
		db.Delete(&newUser)
		return Student{}, "", fmt.Errorf("failed to create student record: %w", err)
	}

	db.Preload("User").Preload("Course").First(&student, student.ID)
	return student, randomPassword, nil
}

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

	// Check if user already exists
	var existingUser user.User
	if err := db.Where("email = ?", tmpUser.Email).First(&existingUser).Error; err == nil {
		return c.Status(402).JSON(fiber.Map{"msg": "user with email " + tmpUser.Email + " already exists"})
	}

	// Hash password if provided
	if tmpUser.Password != "" {
		tmpUser.Password = user.GenerateHash(tmpUser.Password)
	} else {
		// Generate a default password if not provided
		tmpUser.Password = user.GenerateHash("changeme123")
	}

	// Create the user first
	if err := db.Create(&tmpUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to create user: " + err.Error()})
	}

	// Set the user ID and course ID for the student record
	student.UserID = tmpUser.ID
	student.CourseID = student.CourseID

	// Create the student record
	if err := db.Create(&student).Error; err != nil {
		// Rollback: delete the user if student creation fails
		db.Delete(&tmpUser)
		return c.Status(400).JSON(fiber.Map{"msg": "failed to add student: " + err.Error()})
	}

	// Reload student with relations
	db.Preload("User").Preload("Course").First(&student, student.ID)

	return c.Status(201).JSON(fiber.Map{"msg": "student created successfully", "data": student})
}

// CreateForCourse creates a student for a specific course (takes courseId as param)
func CreateForCourse(c *fiber.Ctx, db *gorm.DB) error {
	courseIdParam := c.Params("courseId")
	courseId, err := strconv.ParseUint(courseIdParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid course ID"})
	}

	// Verify course exists
	var courseRecord course.Course
	if err := db.First(&courseRecord, courseId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "course not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find course"})
	}

	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid student details: " + err.Error()})
	}

	student, password, err := createStudentForCourse(db, courseId, req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	if err := SendPasswordEmail(req.Email, req.Name, password); err != nil {
		fmt.Printf("Warning: Failed to send password email to %s: %v\n", req.Email, err)
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

	if err := db.Where("course_id = ?", CourseID).
		Preload("User").
		Preload("Course").
		Preload("Course.Department").
		Preload("Course.Department.Organization").
		Find(&students).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get students"})
	}

	return c.JSON(fiber.Map{"data": students})

}

func CreateBulkForCourse(c *fiber.Ctx, db *gorm.DB) error {
	courseIdParam := c.Params("courseId")
	courseId, err := strconv.ParseUint(courseIdParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid course ID"})
	}

	var courseRecord course.Course
	if err := db.First(&courseRecord, courseId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{"msg": "course not found"})
		}
		return c.Status(400).JSON(fiber.Map{"msg": "failed to find course"})
	}

	var req BulkStudentsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid student details: " + err.Error()})
	}

	if len(req.Students) == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "no students provided"})
	}

	type BulkResult struct {
		Student Student `json:"student"`
		Error   string  `json:"error,omitempty"`
	}

	var successes []Student
	var failures []BulkResult

	for _, studentReq := range req.Students {
		student, password, err := createStudentForCourse(db, courseId, studentReq)
		if err != nil {
			failures = append(failures, BulkResult{Error: fmt.Sprintf("%s (%s)", err.Error(), studentReq.Email)})
			continue
		}

		if err := SendPasswordEmail(studentReq.Email, studentReq.Name, password); err != nil {
			fmt.Printf("Warning: Failed to send password email to %s: %v\n", studentReq.Email, err)
		}

		successes = append(successes, student)
	}

	status := fiber.Map{
		"msg":       "bulk student creation completed",
		"created":   len(successes),
		"failed":    len(failures),
		"successes": successes,
		"errors":    failures,
		"course_id": courseId,
	}

	if len(successes) == 0 {
		return c.Status(400).JSON(status)
	}

	return c.Status(201).JSON(status)
}
