package gormfs

import (
	"io/fs"
	"os"
	"time"
)

var _ fs.FileInfo = (*aferoFile)(nil)

func (af *aferoFile) Mode() os.FileMode {
	panic("aferoFile.Mode not implemented")
}

func (af *aferoFile) ModTime() time.Time {
	panic("aferoFile.ModeTime not implemented")
}

func (af *aferoFile) IsDir() bool {
	panic("aferoFile.IsDir not implemented")
}

func (af *aferoFile) Sys() interface{} { return nil }

func (af *aferoFile) Size() int64 {
	panic("aferoFile.Size not implemented")
}
