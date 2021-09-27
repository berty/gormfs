package gormfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReaddirnames(t *testing.T) {
	fs := TestingFs(t)

	dir := filepath.Join("a", "b", "c")
	subDir := filepath.Join(dir, "d", "e")

	names := []string{"file1", "file2", "file3"}

	require.NoError(t, fs.MkdirAll(subDir, os.ModePerm))

	for _, n := range names {
		p := filepath.Join(dir, n)
		_, err := fs.Create(p)
		require.NoError(t, err)
	}

	f, err := fs.Open(dir)
	require.NoError(t, err)

	readNames, err := f.Readdirnames(-1) // FIXME: implement count
	require.NoError(t, err)

	require.Equal(t, append([]string{"d"}, names...), readNames)
}
