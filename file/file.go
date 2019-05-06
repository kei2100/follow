package file

import "os"

// Open opens the named file for reading and following.
func Open(name string) (*os.File, error) {
	return openFile(name, os.O_RDONLY, 0)
}
