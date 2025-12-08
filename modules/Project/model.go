package project

import (
	"encoding/json"
	"fmt"

	course "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Course"
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Budget struct {
	Currency string
	Value    uint
}

// UnmarshalJSON implements custom JSON unmarshaling for Budget
// Handles both formats:
// 1. Number: 1000 (frontend sends budget as number)
// 2. Object: {"currency": "USD", "value": 1000}
func (b *Budget) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first (frontend format)
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		// It's a number, but we need currency - this will be handled in Create function
		// For now, just set the value and leave currency empty
		b.Value = uint(num)
		b.Currency = ""
		return nil
	}

	// Try to unmarshal as object
	var obj struct {
		Currency string `json:"currency"`
		Value    uint   `json:"value"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		b.Currency = obj.Currency
		b.Value = obj.Value
		return nil
	}

	return fmt.Errorf("cannot unmarshal budget: expected number or object with currency and value")
}

type Project struct {
	gorm.Model
	DepartmentID           int                   `json:"departmentId"`
	CourseID               *uint                 `json:"courseId,omitempty"` // Optional - nullable
	Title                  string                `json:"title"`
	Department             department.Department `json:"department" gorm:"foreignKey:DepartmentID"`
	Course                 *course.Course        `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	Description            string                `json:"description"` // Kept for backward compatibility
	Summary                string                `json:"summary,omitempty" gorm:"type:text"`
	ChallengeStatement     string                `json:"challengeStatement,omitempty" gorm:"type:text"`
	ScopeActivities        string                `json:"scopeActivities,omitempty" gorm:"type:text"`
	DeliverablesMilestones string                `json:"deliverablesMilestones,omitempty" gorm:"type:text"`
	TeamStructure          string                `json:"teamStructure,omitempty"` // "individuals", "groups", "both"
	Duration               string                `json:"duration,omitempty"`      // Project duration (e.g., "12 weeks", "3 months")
	Expectations           string                `json:"expectations,omitempty" gorm:"type:text"`
	Skills                 datatypes.JSON        `json:"skills" gorm:"type:json"`
	Budget                 Budget                `json:"budget" gorm:"embedded;embeddedPrefix:budget_"`
	Deadline               string                `json:"deadline"`
	Capacity               uint                  `json:"capacity" gorm:"default:0"`
	Status                 string                `json:"status" gorm:"default:'pending'"`
	Attachments            datatypes.JSON        `json:"attachments" gorm:"type:json"`
	UserID                 uint                  `json:"userId"`
	User                   user.User             `json:"user" gorm:"foreignKey:UserID"`
	SupervisorID           *uint                 `json:"supervisorId,omitempty"` // Optional - nullable
	Supervisor             *user.User            `json:"supervisor,omitempty" gorm:"foreignKey:SupervisorID"`
}
