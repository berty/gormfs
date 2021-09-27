package gormfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"gorm.io/gorm"
)

// FIXME: handle O_APPEND flag correctly

type aferoFile struct {
	db   *gorm.DB
	name string
	flag int
	head int64
}

var _ afero.File = (*aferoFile)(nil)

func (af *aferoFile) WriteString(s string) (int, error) {
	return af.Write([]byte(s))
}

func (af *aferoFile) WriteAt(p []byte, off int64) (int, error) {
	if af.isReadOnly() {
		return 0, errors.New("file handle is read only")
	}

	f, err := getFile(af.db, af.name)
	if err != nil {
		return 0, err
	}

	newSize := off + int64(len(p))
	if int64(len(f.Data)) < newSize {
		buf := make([]byte, newSize)
		copy(buf, f.Data)
		f.Data = buf
	}

	n := copy(f.Data[off:], p)

	if err := af.db.Save(f).Error; err != nil {
		return 0, err
	}

	return n, nil
}

func (af *aferoFile) Write(p []byte) (int, error) {
	if af.isReadOnly() {
		return 0, errors.New("file handle is read only")
	}

	f, err := getFile(af.db, af.name)
	if err != nil {
		return 0, err
	}

	newSize := af.head + int64(len(p))
	if int64(len(f.Data)) < newSize {
		buf := make([]byte, newSize)
		copy(buf, f.Data)
		f.Data = buf
	}

	n := copy(f.Data[af.head:], p)

	if err := af.db.Save(f).Error; err != nil {
		return 0, err
	}

	af.head += int64(n)

	return n, nil
}

func (af *aferoFile) Truncate(size int64) error {
	if af.isReadOnly() {
		return errors.New("file handle is read only")
	}

	f, err := getFile(af.db, af.name)
	if err != nil {
		return err
	}

	if int64(len(f.Data)) == size {
		return nil
	}

	if int64(len(f.Data)) > size {
		f.Data = f.Data[:size]
	} else {
		buf := make([]byte, size)
		copy(buf, f.Data)
		f.Data = buf
	}

	return af.db.Save(f).Error
}

func (af *aferoFile) Sync() error {
	return nil
}

func (af *aferoFile) Stat() (fs.FileInfo, error) {
	return af, nil
}

func (af *aferoFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		af.head = offset
	case io.SeekCurrent:
		af.head += offset
	case io.SeekEnd:
		af.head = af.Size() + offset
	}
	return af.head, nil
}

func (af *aferoFile) Readdirnames(count int) ([]string, error) {
	return nil, errors.New("aferoFile.Readdirnames not implemented")
}

func (af *aferoFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, errors.New("aferoFile.Readdir not implemented")
}

func (af *aferoFile) ReadAt(p []byte, off int64) (int, error) {
	f, err := getFile(af.db, af.name)
	if err != nil {
		return 0, err
	}

	if off >= int64(len(f.Data)) {
		return 0, io.EOF
	}

	return copy(p, f.Data[off:]), nil
}

func (af *aferoFile) Read(p []byte) (int, error) {
	f, err := getFile(af.db, af.name)
	if err != nil {
		return 0, err
	}

	if af.head >= int64(len(f.Data)) {
		return 0, io.EOF
	}

	n := copy(p, f.Data[af.head:])
	af.head += int64(n)
	return n, nil
}

func (af *aferoFile) Name() string {
	return af.name
}

func (af *aferoFile) Close() error {
	return nil
}

func newAferoFile(db *gorm.DB, name string, flag int) *aferoFile {
	name = filepath.Clean(name)
	return &aferoFile{name: name, db: db, flag: flag}
}

func (af *aferoFile) isReadOnly() bool {
	return af.flag&os.O_RDONLY != 0
}
