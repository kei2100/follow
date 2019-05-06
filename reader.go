package follow

import (
	"fmt"
	"io"
	"os"
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
		positionFile = posfile.InMemory(fileStat, fileInfo.Size())
	}
	if positionFile.FileStat() == nil {
		if err := positionFile.Set(fileStat, fileInfo.Size()); err != nil {
			return errAndClose(err)
		}
	}
	if !stat.SameFile(fileStat, positionFile.FileStat()) {
		logger.Printf("follow: file not found that matches fileStat of the positionFile %+v. reset positionFile.", positionFile.FileStat())
		if err := positionFile.Set(fileStat, fileInfo.Size()); err != nil {
			return errAndClose(err)
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

	return newReader(f, positionFile, opt), nil
}

type reader struct {
	file         *os.File
	positionFile posfile.PositionFile

	watchRotateInterval time.Duration
	detectRotateDelay   time.Duration

	closed  chan struct{}
	rotated <-chan struct{}
}

func newReader(file *os.File, positionFile posfile.PositionFile, opt option) *reader {
	closed := make(chan struct{})
	watchRotateInterval := opt.watchRotateInterval
	detectRotateDelay := opt.detectRotateDelay
	var rotated <-chan struct{}
	if opt.followRotate {
		rotated = watchRotate(closed, file, watchRotateInterval, detectRotateDelay)
	}
	return &reader{
		file:                file,
		positionFile:        positionFile,
		watchRotateInterval: watchRotateInterval,
		detectRotateDelay:   detectRotateDelay,
		closed:              closed,
		rotated:             rotated,
	}
}

// Read reads up to len(b) bytes from the File.
func (r *reader) Read(p []byte) (n int, err error) {
	select {
	default:
		n, err := r.file.Read(p)
		if err != nil {
			return n, err
		}
		if err := r.positionFile.IncreaseOffset(n); err != nil {
			return n, err
		}
		return n, nil

	case <-r.rotated:
		if err := r.file.Close(); err != nil {
			return 0, err
		}
		f, err := file.Open(r.file.Name())
		if err != nil {
			return 0, err
		}
		st, err := stat.Stat(f)
		if err != nil {
			return 0, err
		}
		r.file = f
		if err := r.positionFile.Set(st, 0); err != nil {
			return 0, err
		}
		r.rotated = watchRotate(r.closed, r.file, r.watchRotateInterval, r.detectRotateDelay)
		return r.Read(p)
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

func watchRotate(done chan struct{}, file *os.File, interval, notifyDelay time.Duration) (rotated <-chan struct{}) {
	notify := make(chan struct{})

	go func() {
		tick := time.NewTicker(interval)
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
				currentInfo, err := os.Stat(file.Name())
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					logger.Printf("follow: failed to get current FileStat %s on watchRotate: %+v", file.Name(), err)
					continue
				}
				if !os.SameFile(fileInfo, currentInfo) {
					<-time.After(notifyDelay)
					close(notify)
					return
				}
			}
		}
	}()

	return notify
}
