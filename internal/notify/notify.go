package notify

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/gmbh-micro/defaults"
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
	TAG = defaults.DEFAULT_PROMPT
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

// StdMsg logs a message to stdOut if in verbose mode
func StdMsg(msg string, tab ...int) {
	toStdOutWithTag(checkIndent(tab...) + msg)
}

// StdMsgNoPrompt logs a message to stdOut if in verbose mode
func StdMsgNoPrompt(msg string, tab ...int) {
	toStdOut(checkIndent(tab...) + msg)
}

// StdMsgBlue logs a blue message to stdOut if in verbose mode
func StdMsgBlue(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgBlue, checkIndent(tab...)+msg)
}

// StdMsgBlueNoPrompt logs a blue message to stdOut if in verbose mode
func StdMsgBlueNoPrompt(msg string, tab ...int) {
	toStdOutWithColor(color.FgBlue, checkIndent(tab...)+msg)
}

// StdMsgGreen logs a green message to stdOut if in verbose mode
func StdMsgGreen(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgGreen, checkIndent(tab...)+msg)
}

// StdMsgGreenNoPrompt logs a green message to stdOut if in verbose mode
func StdMsgGreenNoPrompt(msg string, tab ...int) {
	toStdOutWithColor(color.FgGreen, checkIndent(tab...)+msg)
}

// StdMsgMagenta logs a magenta message to stdOut if in verbose mode
func StdMsgMagenta(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgMagenta, checkIndent(tab...)+msg)
}

// StdMsgMagentaNoPrompt logs a magenta message to stdOut if in verbose mode
func StdMsgMagentaNoPrompt(msg string, tab ...int) {
	toStdOutWithColor(color.FgMagenta, checkIndent(tab...)+msg)
}

// StdMsgCyan logs a Cyan message to stdOut if in verbose mode
func StdMsgCyan(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgCyan, checkIndent(tab...)+msg)
}

// StdMsgCyanNoPrompt logs a Cyan message to stdOut if in verbose mode
func StdMsgCyanNoPrompt(msg string, tab ...int) {
	toStdOutWithColor(color.FgCyan, checkIndent(tab...)+msg)
}

// StdMsgErr logs a red error message to stdOut if in verbose mode
func StdMsgErr(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgRed, checkIndent(tab...)+msg)
}

// StdMsgErrNoPrompt logs a red error message to stdOut if in verbose mode
func StdMsgErrNoPrompt(msg string, tab ...int) {
	toStdOutWithColor(color.FgRed, checkIndent(tab...)+msg)
}

// StdMsgDebug logs a highlighted message to stdOut if in verbose mode
func StdMsgDebug(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgYellow, checkIndent(tab...)+msg)
}

// StdMsgLog logs a yellow message to stdOut if in verbose mode
func StdMsgLog(msg string, tab ...int) {
	toStdOutWithColorTag(color.FgYellow, checkIndent(tab...)+msg)
}

func checkIndent(tab ...int) string {
	indent := ""
	if len(tab) >= 1 {
		for i := 0; i <= tab[0]; i++ {
			indent += TAB
		}
	}
	return indent
}

func toStdOutWithTag(msg string) {
	if verbose {
		fmt.Println(TAG + msg)
	}
}

func toStdOut(msg string) {
	if verbose {
		fmt.Println(msg)
	}
}

func toStdOutWithColorTag(c color.Attribute, msg string) {
	if verbose {
		color.Set(c)
		defer color.Unset()
		fmt.Printf(TAG + msg + "\n")
	}
}

func toStdOutWithColor(c color.Attribute, msg string) {
	if verbose {
		color.Set(c)
		defer color.Unset()
		fmt.Printf(msg + "\n")
	}
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
		fmt.Printf(msg + "\n")
		color.Unset()
	}
}
func outB(c color.Attribute, msg string) {
	if verbose {
		color.Set(c).Add(color.Bold)
		fmt.Printf(msg + "\n")
		color.Unset()
	}
}
