package organization

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/gorm"
)

type BillingInformation struct {
}

type Organization struct {
	gorm.Model
	Name       string    `json:"name" gorm:"unique"`
	Type       string    `json:"type"`
	IsApproved bool      `json:"is_approved"`
	UserID     uint      `json:"user_id"`
	User       user.User `json:"user" gorm:"foreignKey:UserID"`
	Logo       string    `json:"logo"`
	BrandColor string    `json:"brand_color"`
}
