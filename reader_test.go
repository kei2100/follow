package follow

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kei2100/follow/internal/testutil"

	"github.com/kei2100/follow/stat"

	"github.com/kei2100/follow/posfile"
)

func TestNoPositionFile(t *testing.T) {
	t.Run("Glow", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		r := mustOpen(f.Name())
		defer r.Close()

		f.WriteString("foo")
		wantRead(t, r, "fo")
		wantPositionFile(t, r.positionFile, fileStat, 2)

		wantRead(t, r, "o")
		wantReadAll(t, r, "")
		wantPositionFile(t, r.positionFile, fileStat, 3)

		f.WriteString("bar")
		wantReadAll(t, r, "bar")
		wantPositionFile(t, r.positionFile, fileStat, 6)
	})

	t.Run("Follow Rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, _ := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpen(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
		defer r.Close()

		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, currentStat := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		wantDetectRotate(t, r, 500*time.Millisecond)
		current.WriteString("foo")
		wantReadAll(t, r, "foo")
		wantPositionFile(t, r.positionFile, currentStat, 3)
	})

	t.Run("No Follow Rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, oldStat := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpen(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0), WithFollowRotate(false))
		defer r.Close()

		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, _ := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		wantNoDetectRotate(t, r, 500*time.Millisecond)
		current.WriteString("foo")
		wantReadAll(t, r, "")
		wantPositionFile(t, r.positionFile, oldStat, 0)
	})

	t.Run("Follow Rotate DetectRotateDelay", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, oldStat := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpen(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(500*time.Millisecond))
		defer r.Close()

		old.WriteString("foo")
		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, currentStat := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		wantReadAll(t, r, "foo")
		wantPositionFile(t, r.positionFile, oldStat, 3)

		wantDetectRotate(t, r, time.Second)
		current.WriteString("barbaz")
		wantReadAll(t, r, "barbaz")
		wantPositionFile(t, r.positionFile, currentStat, 6)
	})
}

func TestWithPositionFile(t *testing.T) {
	t.Run("Works", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		f.WriteString("bar")
		positionFile := posfile.InMemory(fileStat, 2)
		r := mustOpen(f.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "r")
		wantPositionFile(t, r.positionFile, fileStat, 3)

		f.WriteString("baz")
		wantReadAll(t, r, "baz")
		wantPositionFile(t, r.positionFile, fileStat, 6)
	})

	t.Run("Incorrect offset", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		f.WriteString("bar")
		positionFile := posfile.InMemory(fileStat, 4)
		r := mustOpen(f.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "")
		wantPositionFile(t, r.positionFile, fileStat, 3)
	})

	t.Run("Same file not found", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, oldStat := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		old.WriteString("foo")
		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, currentStat := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		positionFile := posfile.InMemory(oldStat, 2)
		r := mustOpen(current.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "")
		wantPositionFile(t, r.positionFile, currentStat, 0)

		current.WriteString("bar")
		wantReadAll(t, r, "bar")
		wantPositionFile(t, r.positionFile, currentStat, 3)
	})
}

func mustOpen(name string, opt ...OptionFunc) *reader {
	r, err := Open(name, opt...)
	if err != nil {
		panic(err)
	}
	return r.(*reader)
}

func wantPositionFile(t *testing.T, positionFile posfile.PositionFile, wantFileStat *stat.FileStat, wantOffset int64) {
	t.Helper()

	if !stat.SameFile(positionFile.FileStat(), wantFileStat) {
		t.Errorf("fileStat not same")
	}
	if g, w := positionFile.Offset(), wantOffset; g != w {
		t.Errorf("offset got %v, want %v", g, w)
	}
}

func wantRead(t *testing.T, reader *reader, want string) {
	t.Helper()

	b := make([]byte, len(want))
	n, err := reader.Read(b)
	if err != nil {
		t.Errorf("failed to read: %v", err)
		return
	}
	if g, w := n, len(b); g != w {
		t.Errorf("nReadBytes got %v, want %v", g, w)
	}
	if g, w := string(b), want; g != w {
		t.Errorf("byteString got %v, want %v", g, w)
	}
}

func wantReadAll(t *testing.T, reader *reader, want string) {
	t.Helper()

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("failed to read all: %v", err)
		return
	}
	if g, w := len(b), len(want); g != w {
		t.Errorf("nReadBytes got %v, want %v", g, w)
	}
	if g, w := string(b), want; g != w {
		t.Errorf("byteString got %v, want %v", g, w)
	}
}

func wantDetectRotate(t *testing.T, reader *reader, timeout time.Duration) {
	t.Helper()

	select {
	case <-reader.rotated:
		return
	case <-time.After(timeout):
		t.Errorf("%s timeout while waiting for detect rotate", timeout)
	}
}

func wantNoDetectRotate(t *testing.T, reader *reader, wait time.Duration) {
	t.Helper()

	select {
	case <-reader.rotated:
		t.Errorf("detect rotate. want not detect")
	case <-time.After(wait):
		return
	}
}
