package log

import (
	"os"
)

// Log is the object that holds onto logging data
type Log struct {
	path       string
	filename   string
	configured bool
	file       *os.File
}

// NewLogFile creates a new log at path with filename name
func NewLogFile(path, filename string) *Log {

	logger := &Log{
		path:     path,
		filename: filename,
	}

	createFilePath(path)
	file := createFile(path + "/" + filename)
	if file != nil {
		logger.configured = true
	}
	logger.file = file

	return logger
}

func (l *Log) Msg(m string) {
	l.file.WriteString(m + "\n")
}

func createFilePath(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

func createFile(pathName string) *os.File {
	file, err := os.OpenFile(pathName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil
	}
	return file
}
