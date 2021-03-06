package gormfs

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
func TestFileDataNameRace(t *testing.T) {
	t.Parallel()
	const someName = "someName"
	const someOtherName = "someOtherName"
	d := FileData{
		name: someName,
	}

	if d.Name() != someName {
		t.Errorf("Failed to read correct Name, was %v", d.Name())
	}

	ChangeFileName(&d, someOtherName)
	if d.Name() != someOtherName {
		t.Errorf("Failed to set Name, was %v", d.Name())
	}

	go func() {
		ChangeFileName(&d, someName)
	}()

	if d.Name() != someName && d.Name() != someOtherName {
		t.Errorf("Failed to read either Name, was %v", d.Name())
	}
}

func TestFileDataModTimeRace(t *testing.T) {
	t.Parallel()
	someTime := time.Now()
	someOtherTime := someTime.Add(1 * time.Minute)

	d := FileData{
		modtime: someTime,
	}

	s := FileInfo{
		FileData: &d,
	}

	if s.ModTime() != someTime {
		t.Errorf("Failed to read correct value, was %v", s.ModTime())
	}

	SetModTime(&d, someOtherTime)
	if s.ModTime() != someOtherTime {
		t.Errorf("Failed to set ModTime, was %v", s.ModTime())
	}

	go func() {
		SetModTime(&d, someTime)
	}()

	if s.ModTime() != someTime && s.ModTime() != someOtherTime {
		t.Errorf("Failed to read either modtime, was %v", s.ModTime())
	}
}

func TestFileDataModeRace(t *testing.T) {
	t.Parallel()
	const someMode = 0777
	const someOtherMode = 0660

	d := FileData{
		mode: someMode,
	}

	s := FileInfo{
		FileData: &d,
	}

	if s.Mode() != someMode {
		t.Errorf("Failed to read correct value, was %v", s.Mode())
	}

	SetMode(&d, someOtherMode)
	if s.Mode() != someOtherMode {
		t.Errorf("Failed to set Mode, was %v", s.Mode())
	}

	go func() {
		SetMode(&d, someMode)
	}()

	if s.Mode() != someMode && s.Mode() != someOtherMode {
		t.Errorf("Failed to read either mode, was %v", s.Mode())
	}
}
*/

// See https://github.com/spf13/afero/issues/286.
func TestFileWriteAt(t *testing.T) {
	t.Parallel()

	gfs := TestingFs(t)
	f, err := gfs.Create("abc.txt")
	require.NoError(t, err)

	testData := []byte{1, 2, 3, 4, 5}
	offset := len(testData)

	// 5 zeros + testdata
	_, err = f.WriteAt(testData, int64(offset))
	require.NoError(t, err)

	// 2 * testdata
	_, err = f.WriteAt(testData, 0)
	require.NoError(t, err)

	// 3 * testdata
	_, err = f.WriteAt(testData, int64(offset*2))
	require.NoError(t, err)

	// 3 * testdata + 5 zeros + testdata
	_, err = f.WriteAt(testData, int64(offset*4))
	require.NoError(t, err)

	// 5 * testdata
	_, err = f.WriteAt(testData, int64(offset*3))
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	expected := bytes.Repeat(testData, 5)
	data, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, expected, data)
}

/*
func TestFileDataIsDirRace(t *testing.T) {
	t.Parallel()

	d := FileData{
		dir: true,
	}

	s := FileInfo{
		FileData: &d,
	}

	if s.IsDir() != true {
		t.Errorf("Failed to read correct value, was %v", s.IsDir())
	}

	go func() {
		s.Lock()
		d.dir = false
		s.Unlock()
	}()

	//just logging the value to trigger a read:
	t.Logf("Value is %v", s.IsDir())
}

func TestFileDataSizeRace(t *testing.T) {
	t.Parallel()

	const someData = "Hello"
	const someOtherDataSize = "Hello World"

	d := FileData{
		data: []byte(someData),
		dir:  false,
	}

	s := FileInfo{
		FileData: &d,
	}

	if s.Size() != int64(len(someData)) {
		t.Errorf("Failed to read correct value, was %v", s.Size())
	}

	go func() {
		s.Lock()
		d.data = []byte(someOtherDataSize)
		s.Unlock()
	}()

	//just logging the value to trigger a read:
	t.Logf("Value is %v", s.Size())

	//Testing the Dir size case
	d.dir = true
	if s.Size() != int64(42) {
		t.Errorf("Failed to read correct value for dir, was %v", s.Size())
	}
}
*/

func TestFileReadAtSeekOffset(t *testing.T) {
	t.Parallel()

	gfs := TestingFs(t)
	f, err := gfs.Create("foo")
	require.NoError(t, err)

	_, err = f.WriteString("TEST")
	require.NoError(t, err)
	offset, err := f.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, int64(0), offset)

	offsetBeforeReadAt, err := f.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	require.Equal(t, int64(0), offsetBeforeReadAt)

	b := make([]byte, 4)
	n, err := f.ReadAt(b, 0)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, "TEST", string(b))

	offsetAfterReadAt, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	if offsetAfterReadAt != offsetBeforeReadAt {
		t.Fatal("ReadAt should not affect offset")
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileWriteAndSeek(t *testing.T) {
	gfs := TestingFs(t)
	f, err := gfs.Create("foo")
	require.NoError(t, err)

	assert := func(expected bool, v ...interface{}) {
		if !expected {
			t.Helper()
			t.Fatal(v...)
		}
	}

	data4 := []byte{0, 1, 2, 3}
	data20 := bytes.Repeat(data4, 5)
	var off int64

	for i := 0; i < 100; i++ {
		// write 20 bytes
		n, err := f.Write(data20)
		assert(err == nil, err)
		off += int64(n)
		assert(n == len(data20), n)
		assert(off == int64((i+1)*len(data20)), off)

		// rewind to start and write 4 bytes there
		cur, err := f.Seek(-off, io.SeekCurrent)
		assert(err == nil, err)
		assert(cur == 0, cur)

		n, err = f.Write(data4)
		assert(err == nil, err)
		assert(n == len(data4), n)

		// back at the end
		cur, err = f.Seek(off-int64(n), io.SeekCurrent)
		assert(err == nil, err)
		assert(cur == off, cur, off)
	}
}
