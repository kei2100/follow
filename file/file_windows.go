package file

import (
	"os"

	"github.com/kei2100/filesharedelete"
)

// Open opens the named file for reading and following.
func Open(name string) (*os.File, error) {
	return filesharedelete.OpenFile(name, os.O_RDONLY, 0)
}
