package chat

import (
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderID uint       `json:"senderId"`
	Sender   user.User  `json:"sender" gorm:"foreignKey:SenderID"`
	Body     string     `json:"body"`
	GroupID  uint       `json:"groupId"`
	Group    user.Group `json:"group" gorm:"foreignKey:GroupID"`
}
