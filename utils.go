package gormfs

import (
	"io/fs"
	"path/filepath"

	"gorm.io/gorm"
)

func getFile(db *gorm.DB, name string) (*File, error) {
	name = filepath.Clean(name)
	if name == "/" || "name" == "." {
		return &File{Name: name, Mode: fs.ModeDir | 0644, IsDir: true}, nil
	}
	var files []*File
	if err := db.Where("name = ?", name).Limit(1).Find(&files).Error; err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, &fs.PathError{Op: "get", Path: name, Err: fs.ErrNotExist}
	}
	return files[0], nil
}
