package gormfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestingFs(t *testing.T) *GormFs {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fs.db")), &gorm.Config{})
	require.NoError(t, err)

	fs, err := NewGormFs(db)
	require.NoError(t, err)

	return fs
}

func TestCreateDelete(t *testing.T) {
	fs := TestingFs(t)

	name := "hello.world"

	for i := 0; i < 42; i++ {
		_, err := fs.Create(name)
		require.NoError(t, err)

		require.NoError(t, fs.Remove(name))
	}
}

func TestCreateDeleteDeep(t *testing.T) {
	fs := TestingFs(t)

	name := "a/b/c/d/e/f/hello.world"
	dir := filepath.Dir(name)

	for i := 0; i < 42; i++ {
		_, err := fs.Create(name)
		require.Error(t, err) // FIXME: check error parity with os

		require.NoError(t, fs.MkdirAll(dir, os.ModePerm))

		_, err = fs.Create(name)
		require.NoError(t, err)

		require.NoError(t, fs.RemoveAll("."))
	}
}

func TestRenameDir(t *testing.T) {
	fs := TestingFs(t)

	dir := filepath.Join("foo", "a", "b", "c")
	child := filepath.Join(dir, "file")

	require.NoError(t, fs.MkdirAll(dir, os.ModePerm))

	_, err := fs.Create(child)
	require.NoError(t, err)

	require.NoError(t, fs.Rename("foo", "bar"))

	_, err = fs.Open(filepath.Join("bar", "a", "b", "c", "file"))
	require.NoError(t, err)
}
