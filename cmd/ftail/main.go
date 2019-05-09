package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kei2100/follow"
)

var (
	positionFilePath string
)

func init() {
	flag.StringVar(&positionFilePath, "position-file", "", "position-file path")
}

func main() {
	flag.Usage = func() {
		command := filepath.Base(os.Args[0])
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n\n", command)
		fmt.Fprintf(out, "  %s [options ...] [file]\n\n", command)
		fmt.Fprintf(out, "The options are as follows:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	subject := flag.Arg(0)
	if subject == "" {
		flag.Usage()
		os.Exit(1)
	}
	pf, err := follow.WithPositionFilePath(positionFilePath)
	if err != nil {
		panic(err)
	}

	r, err := follow.Open(subject, pf)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for {
		select {
		case <-time.After(time.Second):
			_, err := io.Copy(os.Stdout, r)
			if err != nil {
				panic(err)
			}
		}
	}
}
