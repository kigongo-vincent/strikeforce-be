package organization

import (
	"gorm.io/gorm"
)

func FindById(db *gorm.DB, id uint) uint {

	var organization Organization

	if err := db.Where("user_id = ?", id).First(&organization).Error; err != nil {
		return 0
	}

	return organization.ID

}

// FindByIdForAdmin gets organization ID for university-admin or delegated-admin
func FindByIdForAdmin(db *gorm.DB, userID uint, role string) uint {
	if role == "delegated-admin" {
		// For delegated-admin, get org through delegated_accesses table
		var orgID uint
		if err := db.Table("delegated_accesses").
			Where("delegated_user_id = ? AND is_active = ?", userID, true).
			Select("organization_id").
			Scan(&orgID).Error; err != nil {
			return 0
		}
		return orgID
	}
	// For university-admin, use the existing FindById function
	return FindById(db, userID)
}
