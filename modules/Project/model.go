package project

import (
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Budget struct {
	Currency string
	Value    uint
}

type Project struct {
	gorm.Model
	DepartmentID int                   `json:"department_id"`
	Department   department.Department `json:"department" gorm:"foreignKey:DepartmentID"`
	Description  string                `json:"description"`
	Skills       datatypes.JSON        `json:"skills" gorm:"type:json"`
	Budget       Budget                `json:"budget" gorm:"embedded;embeddedPrefix:budget_"`
	Deadline     string                `json:"deadline"`
	Capacity     uint                  `json:"capacity" gorm:"default:0"`
	Status       string                `json:"status" gorm:"default:'pending'"`
	Attachments  datatypes.JSON        `json:"attachments" gorm:"type:json"`
	UserID       uint                  `json:"user_id"`
	User         user.User             `json:"user" gorm:"foreignKey:UserID"`
}
