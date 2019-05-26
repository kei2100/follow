package follow

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kei2100/follow/internal/testutil"
	"github.com/kei2100/follow/posfile"
	"github.com/kei2100/follow/stat"
)

func TestNoPositionFile(t *testing.T) {
	t.Run("Before rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		r := mustOpenReader(f.Name())
		defer r.Close()

		f.WriteString("foo")
		wantRead(t, r, "fo", 10*time.Millisecond, time.Second)
		wantPositionFile(t, r, fileStat, 2)

		wantRead(t, r, "o", 10*time.Millisecond, time.Second)
		wantReadAll(t, r, "")
		wantPositionFile(t, r, fileStat, 3)

		f.WriteString("bar")
		wantReadAll(t, r, "bar")
		wantPositionFile(t, r, fileStat, 6)
	})

	t.Run("After rotate", func(t *testing.T) {
		t.Run("Read old file", func(t *testing.T) {
			t.Run("None old file", func(t *testing.T) {
				t.Parallel()

				td := testutil.CreateTempDir()
				defer td.RemoveAll()

				old, oldStat := td.CreateFile("test.log")
				oldc := testutil.OnceCloser{C: old}
				defer oldc.Close()

				r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
				defer r.Close()

				old.WriteString("old")
				oldc.Close()

				mustRename(old.Name(), old.Name()+".bk")
				mustRemoveFile(old.Name() + ".bk")

				current, currentStat := td.CreateFile(filepath.Base(old.Name()))
				defer current.Close()
				current.WriteString("current")

				wantRead(t, r, "ol", 10*time.Millisecond, time.Second)
				wantPositionFile(t, r, oldStat, 2)
				wantRead(t, r, "d", 10*time.Millisecond, time.Second)
				wantPositionFile(t, r, oldStat, 3)
				wantRead(t, r, "curr", 10*time.Millisecond, time.Second)
				wantPositionFile(t, r, currentStat, 4)
				wantRead(t, r, "ent", 10*time.Millisecond, time.Second)
				wantPositionFile(t, r, currentStat, 7)
			})

			t.Run("Exist old file", func(t *testing.T) {
				t.Run("None remaining bytes", func(t *testing.T) {
					t.Parallel()

					td := testutil.CreateTempDir()
					defer td.RemoveAll()

					old, _ := td.CreateFile("test.log")
					oldc := testutil.OnceCloser{C: old}
					defer oldc.Close()

					r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
					defer r.Close()

					oldc.Close()
					mustRename(old.Name(), old.Name()+".bk")

					current, currentStat := td.CreateFile(filepath.Base(old.Name()))
					defer current.Close()
					current.WriteString("current")

					wantRead(t, r, "c", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, currentStat, 1)
					wantRead(t, r, "urrent", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, currentStat, 7)
				})

				t.Run("Exist remaining bytes", func(t *testing.T) {
					t.Parallel()

					td := testutil.CreateTempDir()
					defer td.RemoveAll()

					old, oldStat := td.CreateFile("test.log")
					oldc := testutil.OnceCloser{C: old}
					defer oldc.Close()

					r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
					defer r.Close()

					old.WriteString("old")
					oldc.Close()
					mustRename(old.Name(), old.Name()+".bk")

					current, currentStat := td.CreateFile(filepath.Base(old.Name()))
					defer current.Close()
					current.WriteString("current")

					wantRead(t, r, "ol", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, oldStat, 2)
					wantRead(t, r, "d", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, oldStat, 3)
					wantRead(t, r, "c", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, currentStat, 1)
					wantRead(t, r, "urrent", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, currentStat, 7)
				})
			})
		})

		t.Run("Read new file", func(t *testing.T) {
			t.Run("None new file", func(t *testing.T) {
				t.Parallel()

				td := testutil.CreateTempDir()
				defer td.RemoveAll()

				f, fileStat := td.CreateFile("test.log")
				fc := testutil.OnceCloser{C: f}
				defer fc.Close()

				r := mustOpenReader(f.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
				defer r.Close()

				f.WriteString("file")
				fc.Close()
				mustRename(f.Name(), f.Name()+".bk")

				wantReadAll(t, r, "file")
				wantPositionFile(t, r, fileStat, 4)
			})

			t.Run("Exists new file", func(t *testing.T) {
				t.Run("Grow", func(t *testing.T) {
					t.Parallel()

					td := testutil.CreateTempDir()
					defer td.RemoveAll()

					old, _ := td.CreateFile("test.log")
					oldc := testutil.OnceCloser{C: old}
					defer oldc.Close()

					r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
					defer r.Close()

					old.WriteString("old")
					oldc.Close()
					mustRename(old.Name(), old.Name()+".bk")

					current, currentStat := td.CreateFile(filepath.Base(old.Name()))
					defer current.Close()
					current.WriteString("current")

					wantRead(t, r, "oldcurrent", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, currentStat, 7)

					current.WriteString("grow")
					wantReadAll(t, r, "grow")
					wantPositionFile(t, r, currentStat, 11)
				})

				t.Run("Rotate Again", func(t *testing.T) {
					t.Parallel()

					td := testutil.CreateTempDir()
					defer td.RemoveAll()

					f1, f1Stat := td.CreateFile("test.log")
					f1c := testutil.OnceCloser{C: f1}
					defer f1c.Close()

					r := mustOpenReader(f1.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
					defer r.Close()

					f1.WriteString("f1")

					wantRead(t, r, "f", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, f1Stat, 1)

					f1c.Close()
					mustRename(f1.Name(), f1.Name()+".bk")

					f2, f2Stat := td.CreateFile(filepath.Base(f1.Name()))
					f2c := testutil.OnceCloser{C: f2}
					defer f2c.Close()
					f2.WriteString("f2")

					wantRead(t, r, "1f", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, f2Stat, 1)

					f2c.Close()
					mustRename(f2.Name(), f2.Name()+".bk")

					f3, f3Stat := td.CreateFile(filepath.Base(f2.Name()))
					defer f3.Close()
					f3.WriteString("f3")

					wantRead(t, r, "2f3", 10*time.Millisecond, time.Second)
					wantPositionFile(t, r, f3Stat, 2)
				})
			})
		})
	})

	t.Run("No Follow Rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, oldStat := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0), WithFollowRotate(false))
		defer r.Close()

		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, _ := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		current.WriteString("foo")
		wantReadAll(t, r, "")
		time.Sleep(100 * time.Millisecond)
		wantReadAll(t, r, "")
		time.Sleep(100 * time.Millisecond)
		wantReadAll(t, r, "")
		time.Sleep(100 * time.Millisecond)
		wantReadAll(t, r, "")
		wantPositionFile(t, r, oldStat, 0)
	})

	t.Run("Follow Rotate DetectRotateDelay", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, oldStat := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(time.Second))
		defer r.Close()

		old.WriteString("foo")
		oldc.Close()
		os.Rename(old.Name(), old.Name()+".bk")
		current, currentStat := td.CreateFile(filepath.Base(old.Name()))
		defer current.Close()

		wantReadAll(t, r, "foo")
		wantPositionFile(t, r, oldStat, 3)

		time.Sleep(100 * time.Millisecond)
		wantReadAll(t, r, "")

		current.WriteString("barbaz")
		wantRead(t, r, "barbaz", 10*time.Millisecond, 3*time.Second)
		wantPositionFile(t, r, currentStat, 6)
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
		r := mustOpenReader(f.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "r")
		wantPositionFile(t, r, fileStat, 3)

		f.WriteString("baz")
		wantReadAll(t, r, "baz")
		wantPositionFile(t, r, fileStat, 6)
	})

	t.Run("Incorrect offset", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		f.WriteString("bar")
		positionFile := posfile.InMemory(fileStat, 4)
		r := mustOpenReader(f.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "")
		wantPositionFile(t, r, fileStat, 3)
	})

	t.Run("Same file not found", func(t *testing.T) {
		t.Run("Rotated file not found", func(t *testing.T) {
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
			r := mustOpenReader(current.Name(), WithPositionFile(positionFile))
			defer r.Close()

			wantReadAll(t, r, "")
			wantPositionFile(t, r, currentStat, 0)

			current.WriteString("bar")
			wantReadAll(t, r, "bar")
			wantPositionFile(t, r, currentStat, 3)
		})

		t.Run("Rotated file found", func(t *testing.T) {
			t.Parallel()

			td := testutil.CreateTempDir()
			defer td.RemoveAll()

			name := "test.log"

			foo, _ := td.CreateFile(name + ".foo-1") // want ignore
			foo.WriteString("foo")
			foo.Close()

			bar, barStat := td.CreateFile(name + ".bar-1") // rotated file
			bar.WriteString("bar")
			bar.Close()

			baz, bazStat := td.CreateFile(name) // current file
			baz.WriteString("baz")
			baz.Close()

			globs := []string{
				filepath.Join(td.Path, name+".bar*"),
				filepath.Join(td.Path, name+".foo*"),
			}
			positionFile := posfile.InMemory(barStat, 1)
			r := mustOpenReader(
				baz.Name(),
				WithPositionFile(positionFile),
				WithRotatedFilePathPatterns(globs),
				WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0),
			)
			defer r.Close()

			wantRead(t, r, "arbaz", 10*time.Millisecond, time.Second)
			wantPositionFile(t, r, bazStat, 3)
		})
	})
}

// TODO multi thread test

func mustOpenReader(name string, opt ...OptionFunc) *Reader {
	r, err := Open(name, opt...)
	if err != nil {
		panic(err)
	}
	return r
}

func mustRemoveFile(name string) {
	if err := os.Remove(name); err != nil {
		panic(err)
	}
}

func mustRename(oldname, newname string) {
	if err := os.Rename(oldname, newname); err != nil {
		panic(err)
	}
}

func wantRead(t *testing.T, r *Reader, want string, interval, timeout time.Duration) {
	t.Helper()

	tick := time.NewTicker(interval)
	defer tick.Stop()
	to := time.After(timeout)
	var buf bytes.Buffer

	for {
		b := make([]byte, len(want)-buf.Len())
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			t.Errorf("failed to read %+v", err)
			return
		}
		if n > 0 {
			buf.Write(b[:n])
		}
		if buf.Len() > len(want) {
			t.Errorf("unexpected read bytes %d, want %s", n, want)
			return
		}
		if buf.Len() == len(want) {
			if g, w := buf.String(), want; g != w {
				t.Errorf("got %v, want %v", g, w)
			}
			return
		}
		select {
		case <-to:
			t.Errorf("timeout %s exceeded. got %s, want %s", timeout, buf.String(), want)
			return
		case <-tick.C:
			continue
		}
	}
}

func wantReadAll(t *testing.T, reader io.Reader, want string) {
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

func wantPositionFile(t *testing.T, r *Reader, wantFileStat *stat.FileStat, wantOffset int64) {
	t.Helper()

	fileStat, offset := r.fu.positionFileInfo()
	if !stat.SameFile(fileStat, wantFileStat) {
		t.Errorf("fileStat not same")
	}
	if g, w := offset, wantOffset; g != w {
		t.Errorf("offset got %v, want %v", g, w)
	}
}
