# follow
[![CircleCI](https://circleci.com/gh/kei2100/follow.svg?style=svg)](https://circleci.com/gh/kei2100/follow)
[![Build status](https://ci.appveyor.com/api/projects/status/yeisq4p3nfghx4j3/branch/master?svg=true)](https://ci.appveyor.com/project/kei2100/follow/branch/master)

A file Reader that behaves like `tail -F`

```go
func ExampleOpen() {
	// create tempfile.
	file, _ := ioutil.TempFile("", "*.log")
	filename := file.Name()

	// create follow.Reader.
	// follow.Reader is a file Reader that behaves like `tail -F`.
	options := []follow.OptionFunc{
		// position-file supported.
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
}
```
