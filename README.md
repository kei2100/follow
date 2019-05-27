# follow
[![CircleCI](https://circleci.com/gh/kei2100/follow.svg?style=svg)](https://circleci.com/gh/kei2100/follow)
[![Build status](https://ci.appveyor.com/api/projects/status/yeisq4p3nfghx4j3/branch/master?svg=true)](https://ci.appveyor.com/project/kei2100/follow/branch/master)

A file Reader that behaves like `tail -F`

```go
func ExampleReader() {
	dir, _ := ioutil.TempDir("", "follow-test")
	path := filepath.Join(dir, "test.log")
	f1, _ := os.Create(path)

	// create follow.Reader.
	// follow.Reader is a file Reader that behaves like `tail -F`
	options := []follow.OptionFunc{
		// position-file supported
		follow.WithPositionFile(posfile.InMemory(nil, 0)),
		follow.WithRotatedFilePathPatterns([]string{filepath.Join(dir, "*.log.*")}),
		follow.WithDetectRotateDelay(0),
		follow.WithWatchRotateInterval(100 * time.Millisecond),
	}
	reader, _ := follow.Open(path, options...)

	f1.WriteString("1")
	b, _ := ioutil.ReadAll(reader)
	fmt.Printf("%s\n", b)

	// rotate
	f1.Close()
	os.Rename(path, path+".f1")
	f2, _ := os.Create(path)

	f2.WriteString("2")
	time.Sleep(500 * time.Millisecond) // wait for detect rotate
	b, _ = ioutil.ReadAll(reader)
	fmt.Printf("%s\n", b)

	// write and rotate while closing the follow.Reader
	reader.Close()
	f2.WriteString("3")
	f2.Close()

	os.Rename(path, path+".f2")
	f3, _ := os.Create(path)
	defer f3.Close()
	f3.WriteString("4")

	// re-open follow.Reader
	reader, _ = follow.Open(path, options...)
	defer reader.Close()

	time.Sleep(500 * time.Millisecond) // wait for detect rotate
	b, _ = ioutil.ReadAll(reader)
	fmt.Printf("%s\n", b)

	// Output:
	// 1
	// 2
	// 34
}
```
