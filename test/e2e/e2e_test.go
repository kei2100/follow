package e2e

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/kei2100/follow"
)

var (
	tempDir   = mkTempDir()
	logName   = "test.log"
	nLogFiles = 5
	nLogLines = 100
	logWait   = 10 * time.Millisecond
)

// TODO refactor

func TestE2E(t *testing.T) {
	t.Parallel()

	defer os.RemoveAll(tempDir)

	logwriter := buildLogWriter()
	logwriterDone := startLogWriter(logwriter)

	stopLogRead := make(chan struct{})
	logReaderDone := make(chan struct{})
	var readResult bytes.Buffer

	go func() {
		defer close(logReaderDone)
		time.Sleep(logWait * 10) // wait for starting the logwriter

		for {
			select {
			case <-stopLogRead:
				return
			default:
				for data := range openAndFollowReadDuration(logWait * time.Duration(nLogLines/2)) {
					readResult.Write(data)
				}
			}
		}
	}()

	<-logwriterDone
	time.Sleep(logWait * 10)
	close(stopLogRead)
	<-logReaderDone

	r := bufio.NewReader(&readResult)
	for i := 0; i < nLogFiles; i++ {
		for j := 0; j < nLogLines; j++ {
			b, _, err := r.ReadLine()
			if err != nil {
				t.Error(err)
			}
			if g, w := string(b), strconv.Itoa(j); g != w {
				t.Errorf("i %v got %v, want %v", i, g, w)
			}
		}
	}
}

func mkTempDir() string {
	dir, err := ioutil.TempDir("", "follow")
	if err != nil {
		panic(err)
	}
	return dir
}

func buildLogWriter() string {
	dstPath := filepath.Join(tempDir, "logwriter"+binExtension())
	srcPath := filepath.Join("testdata", "cmd", "logwriter", "main.go")
	cmd := exec.Command("go", "build", "-o", dstPath, srcPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return dstPath
}

func binExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func startLogWriter(executable string) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		cmd := exec.Command(executable)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = []string{
			"LOG_DIR=" + tempDir,
			"LOG_NAME=" + logName,
			"LOG_FILES=" + strconv.Itoa(nLogFiles),
			"LOG_LINES=" + strconv.Itoa(nLogLines),
			"LOG_WAIT=" + logWait.String(),
		}
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
	}()

	return done
}

func openAndFollowReadDuration(duration time.Duration) <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)
		r, err := follow.Open(filepath.Join(tempDir, logName), followReaderOptions()...)
		if err != nil {
			panic(err)
		}
		defer r.Close()
		timeout := time.After(duration)
		for {
			select {
			case <-timeout:
				return
			default:
				b, err := ioutil.ReadAll(r)
				if err != nil {
					panic(err)
				}
				ch <- b
				time.Sleep(logWait)
			}
		}
	}()

	return ch
}

func followReaderOptions() []follow.OptionFunc {
	pf, err := follow.WithPositionFilePath(filepath.Join(tempDir, "posfile"))
	if err != nil {
		panic(err)
	}
	return []follow.OptionFunc{
		pf,
		follow.WithDetectRotateDelay(logWait * 10 * 2),
		follow.WithReadFromHead(true),
		follow.WithRotatedFilePathPatterns([]string{filepath.Join(tempDir, logName+".*")}),
		follow.WithWatchRotateInterval(logWait),
	}
}
