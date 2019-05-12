package follow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kei2100/follow/stat"

	"github.com/kei2100/follow/file"
	"github.com/kei2100/follow/logger"
	"github.com/kei2100/follow/posfile"
)

// Reader interface.
type Reader interface {
	io.ReadCloser
}

// Open opens the named file and returns the follow.Reader
func Open(name string, opts ...OptionFunc) (Reader, error) {
	opt := option{}
	opt.apply(opts...)

	var f *os.File
	var err error

	errAndClose := func(err error) (Reader, error) {
		if f != nil {
			if cErr := f.Close(); cErr != nil {
				logger.Printf("follow: an error occurred while closing the file %s: %+v", name, cErr)
			}
		}
		if opt.positionFile != nil {
			if cErr := opt.positionFile.Close(); cErr != nil {
				logger.Printf("follow: an error occurred while closing the positionFile: %+v", cErr)
			}
		}
		return nil, err
	}

	f, err = file.Open(name)
	if err != nil {
		return errAndClose(err)
	}
	fileStat, err := stat.Stat(f)
	if err != nil {
		return errAndClose(err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return errAndClose(err)
	}

	positionFile := opt.positionFile
	if positionFile == nil {
		logger.Println("follow: positionFile not specified. use in-memory positionFile.")
		//positionFile = posfile.InMemory(fileStat, fileInfo.Size()) // TODO option?
		positionFile = posfile.InMemory(fileStat, 0)
	}
	if positionFile.FileStat() == nil {
		//if err := positionFile.Set(fileStat, fileInfo.Size()); err != nil { // TODO option?
		if err := positionFile.Set(fileStat, 0); err != nil {
			return errAndClose(err)
		}
	}
	if !stat.SameFile(fileStat, positionFile.FileStat()) {
		logger.Printf("follow: file not found that matches fileStat of the positionFile %+v.", positionFile.FileStat()) // TODO %v
		sameFile, sameFileStat, sameFileInfo, err := findSameFile(opt.rotatedFilePathPatterns, positionFile.FileStat())
		if err != nil {
			if !os.IsNotExist(err) {
				return errAndClose(err)
			}
			logger.Printf("follow: reset positionFile.", positionFile.FileStat())
			//if err := positionFile.Set(fileStat, fileInfo.Size()); err != nil { // TODO option?
			if err := positionFile.Set(fileStat, 0); err != nil {
				return errAndClose(err)
			}
		} else {
			logger.Printf("follow: %s matches fileStat of the positionFile.", sameFile.Name())
			f = sameFile
			fileStat = sameFileStat
			fileInfo = sameFileInfo
		}
	}

	if fileInfo.Size() < positionFile.Offset() {
		// consider file truncated
		logger.Printf("follow: incorrect positionFile offset %d. file size %d. reset offset to %d.", positionFile.Offset(), fileInfo.Size(), fileInfo.Size())
		if err := positionFile.SetOffset(fileInfo.Size()); err != nil {
			return errAndClose(err)
		}
	}
	offset, err := f.Seek(positionFile.Offset(), 0)
	if err != nil {
		return errAndClose(err)
	}
	if offset != positionFile.Offset() {
		return errAndClose(fmt.Errorf("follow: seems like seek failed. positionFile offset %d. file offset %d", positionFile.Offset(), offset))
	}

	return newReader(f, name, positionFile, opt.optionFollowRotate), nil
}

type reader struct {
	file           *os.File
	followFilePath string
	positionFile   posfile.PositionFile
	closed         chan struct{}

	opt                   optionFollowRotate
	rotated               <-chan struct{}
	rotatedRemainingBytes chan int64
}

func newReader(file *os.File, followFilePath string, positionFile posfile.PositionFile, opt optionFollowRotate) *reader {
	closed := make(chan struct{})
	return &reader{
		file:                  file,
		followFilePath:        followFilePath,
		positionFile:          positionFile,
		closed:                closed,
		opt:                   opt,
		rotated:               watchRotate(closed, file, followFilePath, opt),
		rotatedRemainingBytes: make(chan int64, 1),
	}
}

// Read reads up to len(b) bytes from the File.
func (r *reader) Read(p []byte) (n int, err error) {
	select {
	default:
		n, err := r.file.Read(p)
		if err != nil {
			return 0, err
		}
		if err := r.positionFile.IncreaseOffset(n); err != nil {
			return n, err
		}
		return n, nil

	case <-r.rotated:
		r.rotated = nil

		// read remaining bytes from rotated files
		fi, err := r.file.Stat()
		if err != nil {
			return 0, err
		}
		remainingBytes := fi.Size() - r.positionFile.Offset()
		r.rotatedRemainingBytes <- remainingBytes
		return r.Read(p)

	case remainingBytes := <-r.rotatedRemainingBytes:
		n, err := r.Read(p)
		if err != nil && err != io.EOF {
			return n, err
		}
		if err != io.EOF {
			remainingBytes -= int64(n)
			if remainingBytes > 0 {
				r.rotatedRemainingBytes <- remainingBytes
				return n, nil
			}
		}

		// finish read remaining bytes from rotated file
		// open new file
		if err := r.file.Close(); err != nil {
			return n, err
		}
		f, err := file.Open(r.followFilePath)
		if err != nil {
			return n, err
		}
		st, err := stat.Stat(f)
		if err != nil {
			return n, err
		}
		r.file = f
		if err := r.positionFile.Set(st, 0); err != nil {
			return n, err
		}
		r.rotated = watchRotate(r.closed, r.file, r.followFilePath, r.opt)

		switch {
		case len(p) == n:
			return n, nil
		case len(p) > n:
			pp := p[n:]
			nn, err := r.Read(pp)
			return n + nn, err
		default:
			logger.Fatalf("follow: unexpected read bytes size %d and bind bytes size %d", len(p), n)
			return n, nil
		}
	}
}

// Close closes the follow.Reader.
func (r *reader) Close() error {
	if r.closed != nil {
		close(r.closed)
	}
	if err := r.positionFile.Close(); err != nil {
		logger.Printf("follow: an error occurred while closing the positionFile: %+v", err)
	}
	if err := r.file.Close(); err != nil {
		return err
	}
	return nil
}

func findSameFile(globPatterns []string, findStat *stat.FileStat) (*os.File, *stat.FileStat, os.FileInfo, error) {
	var f *os.File
	errAndClose := func(tErr error) (*os.File, *stat.FileStat, os.FileInfo, error) {
		if f != nil {
			if cErr := f.Close(); cErr != nil {
				logger.Printf("follow: an error occurred while closing the file %s: %+v", f.Name(), cErr)
			}
		}
		return nil, nil, nil, tErr
	}

	for _, glob := range globPatterns {
		entries, err := filepath.Glob(glob)
		if err != nil {
			return errAndClose(err)
		}

		for _, ent := range entries {
			f, err = file.Open(ent)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return errAndClose(err)
			}
			fileStat, err := stat.Stat(f)
			if err != nil {
				return errAndClose(err)
			}
			if !stat.SameFile(fileStat, findStat) {
				continue
			}
			// got same file
			fileInfo, err := f.Stat()
			if err != nil {
				return errAndClose(err)
			}
			return f, fileStat, fileInfo, nil
		}
	}
	return nil, nil, nil, os.ErrNotExist
}

func watchRotate(done chan struct{}, file *os.File, followFilePath string, opt optionFollowRotate) (rotated <-chan struct{}) {
	if !opt.followRotate {
		return nil
	}

	notify := make(chan struct{})

	go func() {
		tick := time.NewTicker(opt.watchRotateInterval)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				fileInfo, err := file.Stat()
				if err != nil {
					logger.Printf("follow: failed to get FileStat %s on watchRotate: %+v", file.Name(), err)
					continue
				}
				currentInfo, err := os.Stat(followFilePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					logger.Printf("follow: failed to get current FileStat %s on watchRotate: %+v", followFilePath, err)
					continue
				}
				if !os.SameFile(fileInfo, currentInfo) {
					<-time.After(opt.detectRotateDelay)
					close(notify)
					return
				}
			}
		}
	}()

	return notify
}
