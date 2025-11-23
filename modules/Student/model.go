package student

import (
	course "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Course"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	UserID   uint          `json:"user_id"`
	User     user.User     `json:"user" gorm:"foreignKey:UserID"`
	CourseID uint          `json:"course_id"`
	Course   course.Course `json:"course" gorm:"foreignKey:CourseID"`
}
