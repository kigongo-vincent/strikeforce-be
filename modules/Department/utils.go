package department

import "gorm.io/gorm"

func FindById(db *gorm.DB, id uint) uint {

	var department Department

	if err := db.Where("organization_id = ?", id).First(&department).Error; err != nil {
		return 0
	}

	return department.ID

}
