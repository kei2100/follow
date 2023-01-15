package follow_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kei2100/follow"
	"github.com/kei2100/follow/posfile"
)

func ExampleReader() {
	dir, _ := os.MkdirTemp("", "ExampleReader")
	logpath := filepath.Join(dir, "test.log")
	logfile, _ := os.Create(logpath)
	// Create follow.Reader.
	// follow.Reader is a file Reader that behaves like `tail -F`
	opts := []follow.OptionFunc{
		follow.WithPositionFile(posfile.InMemory(nil, 0)),
		follow.WithRotatedFilePathPatterns([]string{filepath.Join(dir, "test.log.*")}),
		follow.WithDetectRotateDelay(0),
		follow.WithWatchRotateInterval(100 * time.Millisecond),
	}
	reader, _ := follow.Open(logpath, opts...)
	defer reader.Close()
	// Reads log files while tracking their rotation
	go func() {
		for {
			b, _ := io.ReadAll(reader)
			if len(b) > 0 {
				fmt.Print(string(b))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	// Write to logfile
	fmt.Fprintln(logfile, "1")
	fmt.Fprintln(logfile, "2")
	// Rotate logfile
	logfile.Close()
	os.Rename(logpath, logpath+".1")
	logfile, _ = os.Create(logpath)
	// Write to new logfile
	fmt.Fprintln(logfile, "3")
	fmt.Fprintln(logfile, "4")
	logfile.Close()
	time.Sleep(time.Second)

	// Output:
	// 1
	// 2
	// 3
	// 4
}
