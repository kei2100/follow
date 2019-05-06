package posfile

import (
	"path/filepath"
	"testing"

	"github.com/kei2100/follow/internal/testutil"

	"github.com/kei2100/follow/stat"
)

func TestOpenUpdate(t *testing.T) {
	td := testutil.CreateTempDir()
	defer td.RemoveAll()

	file, fileStat := td.CreateFile("foo.log")
	defer file.Close()

	pfpath := filepath.Join(td.Path, "posfile")
	pf, err := Open(pfpath)
	if err != nil {
		t.Fatalf("failed to open posfile: %+v", err)
	}
	pfc := testutil.OnceCloser{C: pf}
	defer pfc.Close()

	pf.Set(fileStat, 0)
	pf.IncreaseOffset(2)

	if !stat.SameFile(pf.FileStat(), fileStat) {
		t.Errorf("not same fileStat\ngot: \n%+v\nwant: \n%+v", pf.FileStat(), fileStat)
	}
	if g, w := pf.Offset(), int64(2); g != w {
		t.Errorf("offset got %v, want %v", g, w)
	}
	if err := pfc.Close(); err != nil {
		t.Fatalf("failed to close: %+v", err)
	}

	pf2, err := Open(pfpath)
	if err != nil {
		t.Fatalf("failed to open posfile: %+v", err)
	}
	pfc2 := testutil.OnceCloser{C: pf2}
	defer pfc2.Close()

	if !stat.SameFile(pf2.FileStat(), fileStat) {
		t.Errorf("not same fileStat\ngot: \n%+v\nwant: \n%+v", pf2.FileStat(), fileStat)
	}
	if g, w := pf2.Offset(), int64(2); g != w {
		t.Errorf("offset got %v, want %v", g, w)
	}
	if err := pfc2.Close(); err != nil {
		t.Fatalf("failed to close: %+v", err)
	}
}
