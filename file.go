package gormfs

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
	"gorm.io/gorm"
)

type aferoFile struct {
	db   *gorm.DB
	name string
}

var _ afero.File = (*aferoFile)(nil)

func (af *aferoFile) WriteString(s string) (int, error) {
	return af.Write([]byte(s))
}

func (af *aferoFile) WriteAt(p []byte, off int64) (int, error) {
	return -1, errors.New("aferoFile.WriteAt not implemented")
}

func (af *aferoFile) Write(p []byte) (int, error) {
	return -1, errors.New("aferoFile.Write not implemented")
}

func (af *aferoFile) Truncate(size int64) error {
	return errors.New("aferoFile.Truncate not implemented")
}

func (af *aferoFile) Sync() error {
	return nil
}

func (af *aferoFile) Stat() (fs.FileInfo, error) {
	return af, nil
}

func (af *aferoFile) Seek(offset int64, whence int) (int64, error) {
	return -1, errors.New("aferoFile.Seek not implemented")
}

func (af *aferoFile) Readdirnames(count int) ([]string, error) {
	return nil, errors.New("aferoFile.Readdirnames not implemented")
}

func (af *aferoFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, errors.New("aferoFile.Readdir not implemented")
}

func (af *aferoFile) ReadAt(p []byte, off int64) (int, error) {
	return -1, errors.New("aferoFile.ReadAt not implemented")
}

func (af *aferoFile) Read(p []byte) (int, error) {
	return -1, errors.New("aferoFile.Read not implemented")
}

func (af *aferoFile) Name() string {
	return af.name
}

func (af *aferoFile) Close() error {
	return nil
}

func newAferoFile(db *gorm.DB, name string) *aferoFile {
	name = filepath.Clean(name)
	return &aferoFile{name: name, db: db}
}
