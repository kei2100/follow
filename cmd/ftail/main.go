package main

import (
	"io"
	"os"
	"time"

	"github.com/kei2100/follow"
)

// TODO opts

func main() {
	subject := os.Args[1]
	posfilePath := os.Args[2]

	pf, err := follow.WithPositionFilePath(posfilePath)
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
