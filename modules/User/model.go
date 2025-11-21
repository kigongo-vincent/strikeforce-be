package user

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Profile struct {
	Avatar   string         `json:"avatar"`
	Bio      string         `json:"bio"`
	Skills   datatypes.JSON `json:"skills" gorm:"type:json"` // JSON field
	Phone    string         `json:"phone"`
	Location string         `json:"location"`
}

type User struct {
	gorm.Model // includes: ID uint, CreatedAt, UpdatedAt, DeletedAt

	Role     string  `json:"role"`
	Email    string  `json:"email" gorm:"unique"`
	Name     string  `json:"name"`
	Password string  `json:"password"`
	Profile  Profile `json:"profile" gorm:"embedded;embeddedPrefix:profile_"`
}
