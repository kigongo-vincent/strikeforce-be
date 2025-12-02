package course

import (
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	Name         string                `json:"name"`
	DepartmentID uint                  `json:"departmentId"`
	Department   department.Department `json:"department" gorm:"foreignKey:DepartmentID"`
}
