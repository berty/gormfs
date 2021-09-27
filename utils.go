package gormfs

import "gorm.io/gorm"

func getFile(db *gorm.DB, name string) (*File, error) {
	var f File
	if err := db.Where("name = ?", name).Limit(1).Find(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}
