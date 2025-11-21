package course

import (
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model
	Name         string                 `json:"name"`
	DepartmentID int                    `json:"department_id"`
	Department   *department.Department `json:"Department" gorm:"foreignKey:DepartmentID"`
}
