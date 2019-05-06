package testutil

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/kei2100/follow/stat"
)

// OnceCloser provides once Close()
type OnceCloser struct {
	once sync.Once
	C    io.Closer
}

// Close once closes C
func (c *OnceCloser) Close() error {
	var err error
	c.once.Do(func() {
		err = c.C.Close()
	})
	return err
}

// Stat return FileStat by name
func Stat(name string) *stat.FileStat {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	s, err := stat.Stat(f)
	if err != nil {
		panic(err)
	}
	return s
}

// CreateTempDir creates a temp dir for testing
func CreateTempDir() *TempDir {
	d, err := ioutil.TempDir("", "follow-test")
	if err != nil {
		panic(err)
	}
	return &TempDir{Path: d}
}

// TempDir for testing
type TempDir struct {
	Path string
}

// RemoveAll removes temp dir and files
func (d *TempDir) RemoveAll() {
	os.RemoveAll(d.Path)
}

// CreateFile creates a file in the temp dir
func (d *TempDir) CreateFile(name string) (*os.File, *stat.FileStat) {
	f, err := os.OpenFile(filepath.Join(d.Path, name), os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0600)
	if err != nil {
		panic(err)
	}
	s, err := stat.Stat(f)
	if err != nil {
		panic(err)
	}
	return f, s
}
