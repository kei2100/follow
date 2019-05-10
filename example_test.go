package follow_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/kei2100/follow"
	"github.com/kei2100/follow/posfile"
)

func ExampleOpen() {
	// create tempfile.
	file, _ := ioutil.TempFile("", "*.log")
	filename := file.Name()

	// create follow.Reader.
	// follow.Reader is a file Reader that behaves like `tail -F`
	options := []follow.OptionFunc{
		// position-file supported
		follow.WithPositionFile(posfile.InMemory(nil, 0)),
		follow.WithRotatedFilePathPatterns([]string{filename + ".*"}),
		follow.WithDetectRotateDelay(0),
	}
	reader, _ := follow.Open(filename, options...)

	// write and read
	file.WriteString("1")
	wantReadString(reader, "1")

	file.WriteString("2")
	wantReadString(reader, "2")

	// rotate
	file.Close()
	os.Rename(filename, filename+".1")

	file, _ = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	file.WriteString("3")
	wantReadString(reader, "3")

	// write and rotate while closing the reader
	reader.Close()
	file.WriteString("4")
	file.Close()
	os.Rename(filename, filename+".2")

	file, _ = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	file.WriteString("5")

	reader, _ = follow.Open(filename, options...)
	wantReadString(reader, "45")

	// Output:
}

func wantReadString(reader io.Reader, want string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var read bytes.Buffer
	for {
		if err := ctx.Err(); err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		read.Write(b)
		if read.Len() >= len(want) {
			break
		}
		time.Sleep(follow.DefaultRotateInterval)
	}
	if read.String() != want {
		log.Fatalf("read %s, want %s", read.String(), want)
	}
}
