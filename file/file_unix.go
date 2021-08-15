//go:build linux || freebsd || darwin
// +build linux freebsd darwin

package file

import "os"

// Open opens the named file for reading and following.
func Open(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_RDONLY, 0)
}
