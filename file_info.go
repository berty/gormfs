package gormfs

import (
	"io/fs"
	"os"
	"time"
)

type FileInfo struct {
	*File
}

var _ fs.FileInfo = (*FileInfo)(nil)

func (fi *FileInfo) Name() string {
	return fi.File.Name
}

func (fi *FileInfo) Mode() os.FileMode {
	return fi.File.Mode
}

func (fi *FileInfo) ModTime() time.Time {
	return fi.File.MTime
}

func (fi *FileInfo) IsDir() bool {
	return fi.File.IsDir
}

func (fi *FileInfo) Sys() interface{} { return nil }

func (fi *FileInfo) Size() int64 {
	return int64(len(fi.File.Data))
}
