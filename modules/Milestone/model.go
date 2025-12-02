package milestone

import (
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	"gorm.io/gorm"
)

type Milestone struct {
	gorm.Model
	ProjectID          uint            `json:"projectId"`
	Project            project.Project `json:"project" gorm:"foreignKey:ProjectID"`
	Title              string          `json:"title"`
	Scope              string          `json:"scope"`
	AcceptanceCriteria string          `json:"acceptanceCriteria"`
	DueDate            string          `json:"dueDate"`
	Amount             int             `json:"amount"`
	Currency           string          `json:"currency"`
	Status             string          `json:"status"`
}
