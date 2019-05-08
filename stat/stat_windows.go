package stat

import (
	"os"
	"syscall"
)

// See
// - https://github.com/golang/go/blob/release-branch.go1.12/src/os/types_windows.go

// porting from os.newFileStatFromGetFileInformationByHandle
func stat(file *os.File) (*FileStat, error) {
	h := syscall.Handle(file.Fd())
	var d syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(h, &d); err != nil {
		return nil, &os.PathError{Op: "GetFileInformationByHandle", Path: file.Name(), Err: err}
	}
	return &FileStat{
		Vol:   d.VolumeSerialNumber,
		IdxHi: d.FileIndexHigh,
		IdxLo: d.FileIndexLow,
	}, nil
}

// FileStat is a os specific file stat
type FileStat struct {
	Vol   uint32
	IdxHi uint32
	IdxLo uint32
}

// porting from os.sameFile
func (s *FileStat) sameFile(other *FileStat) bool {
	return s.Vol == other.Vol && s.IdxHi == other.IdxHi && s.IdxLo == other.IdxLo
}
