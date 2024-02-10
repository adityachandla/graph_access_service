package storage

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type FsImpl struct {
	directory string
}

func InitializeFsService(directory string) Fetcher {
	f, err := os.Open(directory)
	if err != nil {
		panic("Unable to open the directory")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		panic("Unable to fetch file status")
	}
	if !stat.IsDir() {
		panic("Provided path is not a directory")
	}
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}
	return &FsImpl{directory}
}

func (fs *FsImpl) ListFiles() []string {
	//Error checked while initialization
	f, _ := os.Open(fs.directory)
	defer f.Close()
	entries, err := f.ReadDir(0)
	if err != nil {
		panic("Unable to list files")
	}
	res := make([]string, len(entries))
	for i, e := range entries {
		res[i] = fs.directory + e.Name()
	}
	return res
}

func (fs *FsImpl) Fetch(path string, brange ByteRange) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("Unable to open %s %s", path, err))
	}
	defer f.Close()
	f.Seek(0, int(brange.start))
	if brange.end == 0 {
		res, err := io.ReadAll(f)
		if err != nil {
			panic("Error while reading bytes")
		}
		return res
	}
	numBytes := brange.end - brange.start + 1
	res := make([]byte, 0, numBytes)
	for len(res) < cap(res) {
		n, err := f.Read(res[len(res):cap(res)])
		res = res[:len(res)+n]
		if err != nil {
			panic("Error while reading bytes")
		}
	}
	return res
}
