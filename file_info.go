package gormfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type fileInfo struct {
	*File
}

var _ fs.FileInfo = (*fileInfo)(nil)

func (fi *fileInfo) Name() string {
	return filepath.Base(fi.File.Name)
}

func (fi *fileInfo) Mode() os.FileMode {
	return fi.File.Mode
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.File.MTime
}

func (fi *fileInfo) IsDir() bool {
	return fi.File.IsDir
}

func (fi *fileInfo) Sys() interface{} { return nil }

func (fi *fileInfo) Size() int64 {
	return int64(len(fi.File.Data))
}
