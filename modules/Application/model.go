package application

import (
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	ProjectID      uint                  `json:"projectId"`
	Project        project.Project       `json:"project" gorm:"foreignKey:ProjectID"`
	ApplicantType  string                `json:"applicantType" gorm:"default:'INDIVIDUAL'"` // INDIVIDUAL or GROUP
	StudentIDs     datatypes.JSON        `json:"studentIds" gorm:"type:json"`                // Array of user IDs
	GroupID        *uint                 `json:"groupId"`
	Group          *user.Group           `json:"group" gorm:"foreignKey:GroupID"`
	Statement      string                `json:"statement"`
	Status         string                `json:"status" gorm:"default:'SUBMITTED'"` // SUBMITTED, SHORTLISTED, WAITLIST, REJECTED, OFFERED, ACCEPTED, DECLINED, ASSIGNED
	Attachments    datatypes.JSON        `json:"attachments" gorm:"type:json"`      // Array of file paths
	PortfolioScore float64               `json:"portfolioScore" gorm:"default:0"`
	Score          datatypes.JSON        `json:"score" gorm:"type:json"`           // Scoring data: autoScore, manualSupervisorScore, finalScore, skillMatch, portfolioScore, ratingScore, onTimeRate, reworkRate
	OfferExpiresAt *datatypes.Date       `json:"offerExpiresAt"`
}
