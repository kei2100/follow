// +build linux freebsd darwin

package stat

import (
	"fmt"
	"os"
	"syscall"
)

func stat(file *os.File) (*FileStat, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("follow: unexpected FileInfo.Sys() type. name %s, type %T", file.Name(), fi.Sys())
	}
	return &FileStat{Sys: sys}, nil
}

// FileStat is a os specific file stat
type FileStat struct {
	Sys *syscall.Stat_t
}

// See
// - https://github.com/golang/go/blob/release-branch.go1.12/src/os/types_unix.go

// porting from os.sameFile
func (s *FileStat) sameFile(other *FileStat) bool {
	return s.Sys.Dev == other.Sys.Dev && s.Sys.Ino == other.Sys.Ino
}
