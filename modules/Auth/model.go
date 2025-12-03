package auth

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken stores hashed reset tokens issued to users.
type PasswordResetToken struct {
	gorm.Model
	UserID    uint      `json:"userId" gorm:"index"`
	TokenHash string    `json:"-" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expiresAt"`
	UsedAt    *time.Time
}
