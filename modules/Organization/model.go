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
	IsApproved bool      `json:"isApproved"`
	UserID     uint      `json:"userId"`
	User       user.User `json:"user" gorm:"foreignKey:UserID"`
	Website    string    `json:"website"`
	Logo       string    `json:"logo"`
	BrandColor string    `json:"brandColor"`
	Address    string    `json:"address"`
}
