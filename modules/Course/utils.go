package course

import "gorm.io/gorm"

func FindById(db *gorm.DB, id uint) uint {

	var course Course

	if err := db.Where("id = ?", id).First(&course).Error; err != nil {
		return 0
	}

	return course.ID

}

func FindByCreator(db *gorm.DB, id uint) uint {

	var course Course

	if err := db.Where("id = ?", id).First(&course).Error; err != nil {
		return 0
	}

	return course.Department.Organization.UserID

}
