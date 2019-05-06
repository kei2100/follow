package stat_test

import (
	"os"
	"testing"

	"github.com/kei2100/follow/internal/testutil"
	. "github.com/kei2100/follow/stat"
)

func TestSameFile(t *testing.T) {
	td := testutil.CreateTempDir()
	defer td.RemoveAll()

	file, fileStat := td.CreateFile("foo-file")
	file.Close()

	os.Rename(file.Name(), file.Name()+".bk")
	renamedStat := testutil.Stat(file.Name() + ".bk")

	newfile, newfileStat := td.CreateFile("foo-file")
	newfile.Close()

	if !SameFile(fileStat, renamedStat) {
		t.Errorf("stat renamedStat are the not same")
	}
	if SameFile(fileStat, newfileStat) {
		t.Errorf("stat newstat are the same")
	}
}
