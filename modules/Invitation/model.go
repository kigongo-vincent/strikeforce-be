package invitation

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/gorm"
)

type Invitation struct {
	gorm.Model
	Email          string                    `json:"email" gorm:"not null"`
	Name           string                    `json:"name,omitempty"`       // Name provided when creating invitation
	Role           string                    `json:"role" gorm:"not null"` // student, supervisor
	OrganizationID uint                      `json:"organizationId" gorm:"not null"`
	Organization   organization.Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
	DepartmentID   *uint                     `json:"departmentId,omitempty"` // Required for supervisors
	Token          string                    `json:"token" gorm:"uniqueIndex;not null"`
	Status         string                    `json:"status" gorm:"default:'PENDING'"` // PENDING, USED, EXPIRED
	ExpiresAt      time.Time                 `json:"expiresAt" gorm:"not null"`
	UsedAt         *time.Time                `json:"usedAt"`
	UserID         *uint                     `json:"userId"` // Set when invitation is used
	User           *user.User                `json:"user" gorm:"foreignKey:UserID"`
}

// GenerateToken generates a secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
