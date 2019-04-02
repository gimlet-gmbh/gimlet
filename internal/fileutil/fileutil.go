package fileutil

import (
	"io"
	"os"
	"path/filepath"
)

// MkDir @ path
func MkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// CreateFile creates the file at fpath
func CreateFile(fpath string) (*os.File, error) {
	file, err := os.OpenFile(fpath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY|io.SeekStart, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetAbsFpath returns the abolute filepath if it can be optained, else returns
// the same path it was given
func GetAbsFpath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

// FileExists at path?
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetLogFile attempts to add the desired path as an extension to the current
// directory as reported by os.GetWd(). The file is then opened or created
// and returned
func GetLogFile(desiredPathExt, filename string) (*os.File, error) {
	// get pwd
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// make sure that the path extension exists or make the directories needed
	dirPath := filepath.Join(dir, desiredPathExt)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
	}
	// create the file
	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
