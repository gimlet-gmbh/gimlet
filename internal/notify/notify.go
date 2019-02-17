package notify

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
)

// TAB is the amount of indent to set
const TAB = "  "

// SEP is a seperator
const SEP = "-----------------------------------------------"

// CORE is the core logo text
const CORE = `                    _           
  _  ._ _  |_  |_| /   _  ._ _  
 (_| | | | |_) | | \\_ (_) | (/_
  _|                            `

// TAG is the msg to put before a msg
var TAG string
var verbose bool

func init() {
	// TAG = defaults.DEFAULT_PROMPT
	verbose = true
}

// SetVerbose on or off
func SetVerbose(on bool) {
	verbose = on
}

// SetTag changes the tag before a message is printed
func SetTag(tag string) {
	TAG = tag
}

// OpenLogFile at path with filename; will create the path if it does not exists
func OpenLogFile(path, filename string) (*os.File, error) {
	checkDir(path)
	file, err := os.OpenFile(path+filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return nil, errors.New("could not create log file")
	}
	return file, nil
}

func checkDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

// Log is the object that holds onto logging data
type Log struct {
	path     string
	filename string
	verbose  bool
	file     *os.File
	mu       *sync.Mutex
}

// NewLogFile creates a new log at path with filename name
func NewLogFile(path, filename string, verbose bool) *Log {

	logger := &Log{
		path:     path,
		filename: filename,
		verbose:  verbose,
		mu:       &sync.Mutex{},
	}

	createFilePath(path)
	file := createFile(path + "/" + filename)
	logger.file = file

	return logger
}

// Ln writes a message to log
func (l *Log) Ln(format string, a ...interface{}) {
	if l.file != nil {
		l.mu.Lock()
		l.file.WriteString(fmt.Sprintf(format, a...) + "\n")
		l.mu.Unlock()
	}
	if l.verbose {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

// Err writes a message to log
func (l *Log) Err(format string, a ...interface{}) {
	if l.file != nil {
		l.mu.Lock()
		l.file.WriteString(fmt.Sprintf(format, a...) + "\n")
		l.mu.Unlock()
	}
	if l.verbose {
		color.Set(color.FgRed)
		fmt.Println(fmt.Sprintf(format, a...))
		color.Unset()
	}
}

// Sep writes a seperator message to log
func (l *Log) Sep() {
	if l.file != nil {
		l.mu.Lock()
		l.file.WriteString(SEP + "\n")
		l.mu.Unlock()
	}
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////
/// REFACTORED BELOW
///////////////////////////////////////////////////////////////////////////////////////////////////////////

var header = ""

// SetHeader is a string that can be set to precede all writes to stdOut
func SetHeader(s string) {
	if s[len(s)-1] != ' ' {
		header = s + " "
		return
	}
	header = s
}

// LnRedF prints to stdOut a line in red formatted
func LnRedF(format string, a ...interface{}) {
	out(color.FgRed, fmt.Sprintf(format, a...))
}

// LnGreenF prints to stdOut a line in green formatted
func LnGreenF(format string, a ...interface{}) {
	out(color.FgGreen, fmt.Sprintf(format, a...))
}

// LnYellowF prints to stdOut a line in yellow formatted
func LnYellowF(format string, a ...interface{}) {
	out(color.FgYellow, fmt.Sprintf(format, a...))
}

// LnBlueF prints to stdOut a line in blue formatted
func LnBlueF(format string, a ...interface{}) {
	out(color.FgBlue, fmt.Sprintf(format, a...))
}

// LnMagentaF prints to stdOut a line in magenta formatted
func LnMagentaF(format string, a ...interface{}) {
	out(color.FgMagenta, fmt.Sprintf(format, a...))
}

// LnCyanF prints to stdOut a line in cyan formatted
func LnCyanF(format string, a ...interface{}) {
	out(color.FgCyan, fmt.Sprintf(format, a...))
}

// LnWhiteF prints to stdOut a line in white formatted
func LnWhiteF(format string, a ...interface{}) {
	out(color.FgWhite, fmt.Sprintf(format, a...))
}

// LnBRedF prints to stdOut a line in bold red formatted
func LnBRedF(format string, a ...interface{}) {
	outB(color.FgRed, fmt.Sprintf(format, a...))
}

// LnBGreenF prints to stdOut a line in bold green formatted
func LnBGreenF(format string, a ...interface{}) {
	outB(color.FgGreen, fmt.Sprintf(format, a...))
}

// LnBYellowF prints to stdOut a line in bold yellow formatted
func LnBYellowF(format string, a ...interface{}) {
	outB(color.FgYellow, fmt.Sprintf(format, a...))
}

// LnBBlueF prints to stdOut a line in bold blue formatted
func LnBBlueF(format string, a ...interface{}) {
	outB(color.FgBlue, fmt.Sprintf(format, a...))
}

// LnBMagentaF prints to stdOut a line in bold magenta formatted
func LnBMagentaF(format string, a ...interface{}) {
	outB(color.FgMagenta, fmt.Sprintf(format, a...))
}

// LnBCyanF prints to stdOut a line in bold cyan formatted
func LnBCyanF(format string, a ...interface{}) {
	outB(color.FgCyan, fmt.Sprintf(format, a...))
}

// LnBWhiteF prints to stdOut a line in bold white formatted
func LnBWhiteF(format string, a ...interface{}) {
	outB(color.FgWhite, fmt.Sprintf(format, a...))
}

func out(c color.Attribute, msg string) {
	if verbose {
		color.Set(c)
		fmt.Printf(header + msg + "\n")
		color.Unset()
	}
}
func outB(c color.Attribute, msg string) {
	if verbose {
		color.Set(c).Add(color.Bold)
		fmt.Printf(header + msg + "\n")
		color.Unset()
	}
}

/**********************************************************************************
**** OS Helpers
**********************************************************************************/

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
		os.Mkdir(dir, 0755)
	}
	// create the file
	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetLogFileWithPath attempts to add the desired path as an extension to the current
// directory as reported by os.GetWd(). The file is then opened or created
// and returned
func GetLogFileWithPath(path, filename string) (*os.File, error) {
	// make sure that the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	// create the file
	filePath := filepath.Join(path, filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Getpwd returns the directory that the process was launched from according to the os package
// Unlike the os package it never returns and error, only an empty string
func Getpwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// notify.LnBCyanF("                      __                                        ")
// notify.LnBCyanF("  _  ._ _  |_  |_|   (_   _  ._   o  _  _  |_|  _  | ._   _  ._ ")
// notify.LnBCyanF(" (_| | | | |_) | |   __) (/_ | \\/ | (_ (/_ | | (/_ | |_) (/_ |  ")
// notify.LnBCyanF("  _|                                                 |          ")
