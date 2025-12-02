package department

import (
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	"gorm.io/gorm"
)

type Department struct {
	gorm.Model
	Name           string                    `json:"name"`
	OrganizationID uint                      `json:"organizationId"`
	Organization   organization.Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
}
