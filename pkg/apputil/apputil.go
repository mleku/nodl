package apputil

import (
	"os"
	"path/filepath"
)

// EnsureDir checks a file could be written to a path, creates the directories
// as needed
func EnsureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); chk.E(serr) {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if chk.E(merr) {
			panic(merr)
		}
	}
}

// FileExists reports whether the named file or directory exists.
func FileExists(filePath string) bool {
	_, e := os.Stat(filePath)
	return !chk.E(e)
}
