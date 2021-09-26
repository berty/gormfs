package gormfs

import (
	"io/fs"
	"time"
)

type File struct {
	Name  string `gorm:"primaryKey"`
	Mode  fs.FileMode
	ATime time.Time
	MTime time.Time
	IsDir bool
}

var allModels = []interface{}{&File{}}
