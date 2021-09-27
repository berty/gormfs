package gormfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

// FIXME: set File.Mode, File.ATime, File.User and File.Group correctly

type GormFs struct {
	db *gorm.DB
}

func NewGormFs(db *gorm.DB) (*GormFs, error) {
	if err := db.AutoMigrate(allModels...); err != nil {
		return nil, errors.Wrap(err, "migrate db")
	}
	return &GormFs{db: db}, nil
}

var _ afero.Fs = (*GormFs)(nil)

func (fs *GormFs) Chmod(name string, mode os.FileMode) error {
	f, err := fs.get(name)
	if err != nil {
		return err
	}
	f.Mode = mode
	f.MTime = time.Now()
	return fs.db.Save(f).Error
}

func (fs *GormFs) Chown(name string, uid, gid int) error {
	f, err := fs.get(name)
	if err != nil {
		return err
	}
	f.User = uid
	f.Group = gid
	f.MTime = time.Now()
	return fs.db.Save(f).Error
}

func (fs *GormFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f, err := fs.get(name)
	if err != nil {
		return err
	}
	f.ATime = atime
	f.MTime = mtime
	return fs.db.Save(f).Error
}

func (fs *GormFs) Create(name string) (afero.File, error) {
	if !fs.hasParent(name) {
		return nil, fmt.Errorf("parent of %s does not exist", name) // FIXME: error parity with os
	}
	now := time.Now()
	if err := fs.db.Create(&File{Name: filepath.Clean(name), ATime: now, MTime: now}).Error; err != nil {
		return nil, errors.Wrap(err, "create db file")
	}
	return fs.OpenFile(name, os.O_RDWR, os.ModePerm)
}

func (fs *GormFs) Mkdir(name string, perm os.FileMode) error {
	if !fs.hasParent(name) {
		return fmt.Errorf("parent of %s does not exist", name) // FIXME: error parity with os
	}
	if err := fs.db.Create(&File{Name: filepath.Clean(name), IsDir: true, Mode: perm}).Error; err != nil {
		return err
	}
	return nil
}

func (fs *GormFs) MkdirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	paths := strings.Split(path, "/") // FIXME: breaks on non-unix
	for i, elem := range paths {
		name := filepath.Join(append(paths[:i], elem)...)
		if err := fs.Mkdir(name, perm); err != nil {
			return err
		}
	}
	return nil
}

func (fs *GormFs) Name() string {
	return "GormFs"
}

func (fs *GormFs) Open(name string) (afero.File, error) {
	return newAferoFile(fs.db, name, os.O_RDONLY), nil
}

func (fs *GormFs) OpenFile(name string, flag int, perm fs.FileMode) (afero.File, error) {
	if !fs.exists(name) {
		if flag&os.O_CREATE == 0 {
			return nil, errors.New("no such file or directory") // FIXME: error parity with os
		}
		if err := fs.db.Create(&File{Name: filepath.Clean(name), Mode: perm}).Error; err != nil {
			return nil, err
		}
	}
	return newAferoFile(fs.db, name, flag), nil
}

func (fs *GormFs) Remove(name string) error {
	return fs.db.Delete(&File{Name: filepath.Clean(name)}).Error
}

func (fs *GormFs) RemoveAll(path string) error {
	path = filepath.Clean(path)
	return fs.db.
		Where("name LIKE ?", filepath.Join(path, "%")).Or("name = ? AND is_dir = true", path). // FIXME: support paths with %
		Delete(&File{}).Error
}

func (fs *GormFs) Rename(oldname, newname string) error {
	oldname = filepath.Clean(oldname)
	newname = filepath.Clean(newname)

	if !fs.exists(oldname) {
		return errors.New("no such file or directory") // FIXME: error parity with os
	}

	oldFiles := []*File{}
	if err := fs.db.Where("name LIKE ?", filepath.Join(oldname, "%")).Or("name = ?", oldname).Find(&oldFiles).Error; err != nil {
		return errors.Wrap(err, "find files")
	}

	now := time.Now()

	newFiles := make([]*File, len(oldFiles))
	for i := range oldFiles {
		f := *oldFiles[i]
		newFiles[i] = &f
		fnn := strings.TrimPrefix(oldFiles[i].Name, oldname)
		newFiles[i].Name = newname + fnn
		newFiles[i].MTime = now
		fmt.Printf("renamed %s to %s\n", oldFiles[i].Name, newFiles[i].Name)
	}

	if err := fs.db.Save(newFiles).Error; err != nil {
		return errors.Wrap(err, "save files")
	}

	if err := fs.db.Delete(oldFiles).Error; err != nil {
		return errors.Wrap(err, "delete rename remains")
	}

	return nil
}

func (fs *GormFs) Stat(name string) (fs.FileInfo, error) {
	return newAferoFile(fs.db, name, os.O_RDONLY).Stat()
}

func (fs *GormFs) hasParent(name string) bool {
	name = filepath.Clean(name)
	parent := filepath.Dir(name)
	if parent == "." || parent == "/" { // FIXME: breaks on non-unix
		return true
	}
	return fs.db.Where("name = ? AND is_dir = true", parent).Limit(1).Find(&File{}).RowsAffected != 0
}

func (fs *GormFs) exists(name string) bool {
	name = filepath.Clean(name)
	if name == "." || name == "/" { // FIXME: breaks on non-unix
		return true
	}
	return fs.db.Where("name = ?", name).Limit(1).Find(&File{}).RowsAffected != 0
}

func (fs *GormFs) get(name string) (*File, error) {
	var f File
	if err := fs.db.Where("name = ?", name).Limit(1).Find(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}
