package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	dir := os.Getenv("LOG_DIR")
	name := os.Getenv("LOG_NAME")
	nFiles, _ := strconv.Atoi(os.Getenv("LOG_FILES"))
	nLines, _ := strconv.Atoi(os.Getenv("LOG_LINES"))
	wait, _ := time.ParseDuration(os.Getenv("LOG_WAIT"))

	if dir == "" {
		d, err := ioutil.TempDir("", "follow")
		if err != nil {
			panic(err)
		}
		dir = d
	}
	if name == "" {
		name = "test.log"
	}
	if nFiles == 0 {
		nFiles = 5
	}
	if nLines == 0 {
		nLines = 100
	}
	if nLines <= 10 {
		panic("LOG_LINES must greater than 10")
	}
	if wait == 0 {
		wait = time.Millisecond
	}

	logFile := openLogFile(filepath.Join(dir, name))
	defer logFile.Close()

	for i := 0; i < nFiles; i++ {
		writeNLines(logFile, 0, nLines-10, wait)
		if !(i+1 < nFiles) { // last loop
			writeNLines(logFile, nLines-10, 10, wait)
			break
		}
		shiftRotatedLogFiles(dir, name, nFiles)
		currentLogFile, rotatedLogFile := rotateLogFile(logFile)
		writeNLines(rotatedLogFile, nLines-10, 10, wait)
		rotatedLogFile.Close()
		logFile = currentLogFile
	}
}

func openLogFile(name string) *os.File {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

func shiftRotatedLogFiles(dir, prefix string, maxLogFiles int) {
	path := filepath.Join(dir, prefix)
	for i := maxLogFiles; i > 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", path, i-1), fmt.Sprintf("%s.%d", path, i))
	}
}

func rotateLogFile(logFile *os.File) (current, rotated *os.File) {
	logFile.Close() // fow Windows

	currentName := logFile.Name()
	rotatedName := fmt.Sprintf("%s.%d", logFile.Name(), 1)
	os.Rename(currentName, rotatedName)

	return openLogFile(currentName), openLogFile(rotatedName)
}

func writeNLines(f *os.File, start, nLines int, wait time.Duration) {
	for i := 0; i < nLines; i++ {
		fmt.Fprintln(f, strconv.Itoa(start+i))
		time.Sleep(wait)
	}
}
