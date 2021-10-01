package gormfs

import (
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
// FIXME: handle flag correctly

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

func (f *GormFs) Chmod(name string, mode fs.FileMode) error {
	file, err := getFile(f.db, name)
	if err != nil {
		return err
	}
	isDir := file.Mode&fs.ModeDir != 0
	file.Mode = mode
	if isDir {
		file.Mode |= fs.ModeDir
	}
	file.MTime = time.Now()
	return f.db.Save(file).Error
}

func (f *GormFs) Chown(name string, uid, gid int) error {
	file, err := getFile(f.db, name)
	if err != nil {
		return err
	}
	file.User = uid
	file.Group = gid
	file.MTime = time.Now()
	return f.db.Save(file).Error
}

func (f *GormFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	file, err := getFile(f.db, name)
	if err != nil {
		return err
	}
	file.ATime = atime
	file.MTime = mtime
	return f.db.Save(file).Error
}

func (f *GormFs) Create(name string) (afero.File, error) {
	if !f.hasParent(name) {
		return nil, &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrNotExist}
	}
	now := time.Now()
	if err := f.db.Create(&File{Name: filepath.Clean(name), ATime: now, MTime: now}).Error; err != nil {
		return nil, errors.Wrap(err, "create db file")
	}
	return f.OpenFile(name, os.O_RDWR, os.ModePerm)
}

func (f *GormFs) Mkdir(name string, perm fs.FileMode) error {
	name = filepath.Clean(name)
	if f.exists(name) {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
	}
	if !f.hasParent(name) {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrNotExist}
	}
	if err := f.db.Create(&File{Name: name, IsDir: true, Mode: perm | fs.ModeDir}).Error; err != nil {
		return err
	}
	return nil
}

func (f *GormFs) MkdirAll(path string, perm fs.FileMode) error {
	path = filepath.Clean(path)
	paths := strings.Split(path, "/") // FIXME: breaks on non-unix
	if len(paths) > 0 && paths[0] == "" {
		paths[0] = "/"
	}
	for i, elem := range paths {
		name := filepath.Join(append(paths[:i], elem)...)
		if f.exists(name) {
			continue
		}
		if err := f.Mkdir(name, perm); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func (f *GormFs) Name() string {
	return "GormFs"
}

func (f *GormFs) Open(name string) (afero.File, error) {
	return f.OpenFile(name, os.O_RDONLY, 0)
}

func (f *GormFs) OpenFile(name string, flag int, perm fs.FileMode) (afero.File, error) {
	name = filepath.Clean(name)
	if f.exists(name) {
		if flag&os.O_CREATE != 0 && flag&os.O_EXCL != 0 {
			return nil, &fs.PathError{Op: "openf", Path: name, Err: fs.ErrExist}
		}
	} else {
		if flag&os.O_CREATE == 0 {
			return nil, &fs.PathError{Op: "openf", Path: name, Err: fs.ErrNotExist}
		}
		if err := f.db.Create(&File{Name: filepath.Clean(name), Mode: perm}).Error; err != nil {
			return nil, err
		}
	}
	return newAferoFile(f.db, name, flag)
}

func (f *GormFs) Remove(name string) error {
	name = filepath.Clean(name)
	if !f.exists(name) {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
	}
	return f.db.Delete(&File{Name: name}).Error
}

func (f *GormFs) RemoveAll(path string) error {
	path = filepath.Clean(path)
	return f.db.
		Where("name LIKE ?", filepath.Join(path, "%")).Or("name = ? AND is_dir = true", path). // FIXME: support paths with %
		Delete(&File{}).Error
}

func (f *GormFs) Rename(oldname, newname string) error {
	oldname = filepath.Clean(oldname)
	newname = filepath.Clean(newname)

	if !f.exists(oldname) {
		return &fs.PathError{Op: "rename", Path: oldname, Err: fs.ErrNotExist} // FIXME: error parity with os
	}

	oldFiles := []*File{}
	if err := f.db.Where("name LIKE ?", filepath.Join(oldname, "%")).Or("name = ?", oldname).Find(&oldFiles).Error; err != nil {
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
	}

	if err := f.db.Save(newFiles).Error; err != nil {
		return errors.Wrap(err, "save files")
	}

	if err := f.db.Delete(oldFiles).Error; err != nil {
		return errors.Wrap(err, "delete rename remains")
	}

	return nil
}

func (f *GormFs) Stat(name string) (fs.FileInfo, error) {
	file, err := newAferoFile(f.db, name, os.O_RDONLY)
	if err != nil {
		return nil, err
	}

	return file.Stat()
}

func (f *GormFs) hasParent(name string) bool {
	name = filepath.Clean(name)
	parent := filepath.Dir(name)
	if parent == "." || parent == "/" { // FIXME: breaks on non-unix
		return true
	}
	return f.db.Where("name = ? AND is_dir = true", parent).Limit(1).Find(&File{}).RowsAffected != 0
}

func (f *GormFs) exists(name string) bool {
	name = filepath.Clean(name)
	if name == "." || name == "/" { // FIXME: breaks on non-unix
		return true
	}
	return f.db.Where("name = ?", name).Limit(1).Find(&File{}).RowsAffected != 0
}
