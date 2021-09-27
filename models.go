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
	User  int
	Group int
	Data  []byte
}

var allModels = []interface{}{&File{}}
