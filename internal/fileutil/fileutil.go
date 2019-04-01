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

// CopyFile from sname to dname
func CopyFile(sname, dname string) error {

	src, err := os.Open(sname)
	if err != nil {
		return err
	}

	dst, err := os.Create(dname)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	srcinfo, err := os.Stat(sname)
	if err != nil {
		return err
	}
	err = os.Chmod(dname, srcinfo.Mode())
	if err != nil {
		return err
	}
	return nil
}
