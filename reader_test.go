package follow

import (
	"io"
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
	t.Run("Before rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		f, fileStat := td.CreateFile("test.log")
		defer f.Close()

		r := mustOpenReader(f.Name())
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

				wantDetectRotate(t, r, 500*time.Millisecond)
				wantRead(t, r, "ol")
				wantPositionFile(t, r.positionFile, oldStat, 2)
				wantRead(t, r, "d")
				wantPositionFile(t, r.positionFile, currentStat, 0)
				wantReadAll(t, r, "current")
				wantPositionFile(t, r.positionFile, currentStat, 7)
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

					wantDetectRotate(t, r, 500*time.Millisecond)
					wantRead(t, r, "c")
					wantPositionFile(t, r.positionFile, currentStat, 1)
					wantReadAll(t, r, "urrent")
					wantPositionFile(t, r.positionFile, currentStat, 7)
				})

				t.Run("Exist remaining bytes", func(t *testing.T) {
					t.Run("Fixed renaming bytes", func(t *testing.T) {
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

						wantDetectRotate(t, r, 500*time.Millisecond)
						wantRead(t, r, "ol")
						wantPositionFile(t, r.positionFile, oldStat, 2)
						wantRead(t, r, "d")
						wantPositionFile(t, r.positionFile, currentStat, 0)
						wantReadAll(t, r, "current")
						wantPositionFile(t, r.positionFile, currentStat, 7)
					})

					t.Run("Increase remaining bytes", func(t *testing.T) {
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

						wantDetectRotate(t, r, 500*time.Millisecond)
						wantRead(t, r, "ol")
						wantPositionFile(t, r.positionFile, oldStat, 2)

						// increase
						old = mustOpenfile(old.Name()+".bk", os.O_APPEND|os.O_WRONLY)
						defer old.Close()
						old.WriteString("maybeIgnored")

						wantRead(t, r, "d")
						wantPositionFile(t, r.positionFile, currentStat, 0)
						wantReadAll(t, r, "current")
						wantPositionFile(t, r.positionFile, currentStat, 7)
					})

					t.Run("Decrease remaining bytes", func(t *testing.T) {
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

						wantDetectRotate(t, r, 500*time.Millisecond)
						wantRead(t, r, "ol")
						wantPositionFile(t, r.positionFile, oldStat, 2)

						// decrease
						old = mustOpenfile(old.Name()+".bk", os.O_TRUNC|os.O_WRONLY)
						defer old.Close()
						old.WriteString("ol")

						wantRead(t, r, "c")
						wantPositionFile(t, r.positionFile, currentStat, 1)
						wantReadAll(t, r, "urrent")
						wantPositionFile(t, r.positionFile, currentStat, 7)
					})
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

				wantNoDetectRotate(t, r, 500*time.Millisecond)
				wantReadAll(t, r, "file")
				wantPositionFile(t, r.positionFile, fileStat, 4)
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

					wantDetectRotate(t, r, 500*time.Millisecond)
					wantReadAll(t, r, "oldcurrent")
					wantPositionFile(t, r.positionFile, currentStat, 7)

					current.WriteString("grow")
					wantReadAll(t, r, "grow")
					wantPositionFile(t, r.positionFile, currentStat, 11)
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
					f1c.Close()
					mustRename(f1.Name(), f1.Name()+".bk")

					f2, f2Stat := td.CreateFile(filepath.Base(f1.Name()))
					f2c := testutil.OnceCloser{C: f2}
					defer f2c.Close()

					wantDetectRotate(t, r, 500*time.Millisecond)
					wantRead(t, r, "f")
					wantPositionFile(t, r.positionFile, f1Stat, 1)
					wantRead(t, r, "1")
					wantPositionFile(t, r.positionFile, f2Stat, 0)

					f2.WriteString("f2")
					f2c.Close()
					mustRename(f2.Name(), f2.Name()+".bk")

					f3, f3Stat := td.CreateFile(filepath.Base(f2.Name()))
					defer f3.Close()
					f3.WriteString("f3")

					wantDetectRotate(t, r, 500*time.Millisecond)
					wantRead(t, r, "f")
					wantPositionFile(t, r.positionFile, f2Stat, 1)
					wantRead(t, r, "2")
					wantPositionFile(t, r.positionFile, f3Stat, 0)
					wantRead(t, r, "f3")
					wantPositionFile(t, r.positionFile, f3Stat, 2)
				})
			})
		})
	})

	t.Run("Follow Rotate", func(t *testing.T) {
		t.Parallel()

		td := testutil.CreateTempDir()
		defer td.RemoveAll()

		old, _ := td.CreateFile("test.log")
		oldc := testutil.OnceCloser{C: old}
		defer oldc.Close()

		r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0))
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

		r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(0), WithFollowRotate(false))
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

		r := mustOpenReader(old.Name(), WithWatchRotateInterval(10*time.Millisecond), WithDetectRotateDelay(500*time.Millisecond))
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
		r := mustOpenReader(f.Name(), WithPositionFile(positionFile))
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
		r := mustOpenReader(f.Name(), WithPositionFile(positionFile))
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
		r := mustOpenReader(current.Name(), WithPositionFile(positionFile))
		defer r.Close()

		wantReadAll(t, r, "")
		wantPositionFile(t, r.positionFile, currentStat, 0)

		current.WriteString("bar")
		wantReadAll(t, r, "bar")
		wantPositionFile(t, r.positionFile, currentStat, 3)
	})
}

// TODO multi goroutine test

func mustOpenReader(name string, opt ...OptionFunc) *reader {
	r, err := Open(name, opt...)
	if err != nil {
		panic(err)
	}
	return r.(*reader)
}

func mustOpenfile(name string, flag int) *os.File {
	f, err := os.OpenFile(name, flag, 0600)
	if err != nil {
		panic(err)
	}
	return f
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

	if want == "" {
		return
	}

	var got string
	for len(got) != len(want) {
		b := make([]byte, len(want)-len(got))
		n, err := reader.Read(b)
		if err != nil {
			if err != io.EOF {
				t.Errorf("failed to read: %v", err)
			}
			break
		}
		got += string(b[:n])
	}

	if g, w := got, want; g != w {
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
