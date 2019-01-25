package notify

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// TAB is the amount of indent to set
const TAB = "  "

// TAG is the msg to put before a msg
var TAG string
var lvl int
var verbose bool

func init() {
	TAG = "[gmbh] "
	lvl = 0
	verbose = true
}

// SetLevel of the logging
func SetLevel(newLvl int) {
	lvl = newLvl
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
	if verbose {
		toStdOutWithTag(checkIndent(tab...) + msg)
	}
}

// StdMsgBlue logs a blue message to stdOut if in verbose mode
func StdMsgBlue(msg string, tab ...int) {
	if verbose {
		toStdOutWithColorTag(color.FgBlue, checkIndent(tab...)+msg)
	}
}

// StdMsgGreen logs a green message to stdOut if in verbose mode
func StdMsgGreen(msg string, tab ...int) {
	if verbose {
		toStdOutWithColorTag(color.FgGreen, checkIndent(tab...)+msg)
	}
}

// StdMsgMagenta logs a magenta message to stdOut if in verbose mode
func StdMsgMagenta(msg string, tab ...int) {
	if verbose {
		toStdOutWithColorTag(color.FgMagenta, checkIndent(tab...)+msg)
	}
}

// StdMsgErr logs a red error message to stdOut if in verbose mode
func StdMsgErr(msg string, tab ...int) {
	if verbose {
		toStdOutWithColorTag(color.FgRed, checkIndent(tab...)+msg)
	}
}

// StdMsgDebug logs a highlighted message to stdOut if in verbose mode
func StdMsgDebug(msg string, tab ...int) {
	if verbose {
		// outMsg := checkIndent(tab...)+msg
		toStdOutWithColorTag(color.BgCyan, checkIndent(tab...)+msg)
	}
}

// StdMsgLog logs a yellow message to stdOut if in verbose mode
func StdMsgLog(msg string, tab ...int) {
	if verbose {
		toStdOutWithColorTag(color.FgYellow, checkIndent(tab...)+msg)
	}
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
	fmt.Println(TAG + msg)
}

func toStdOutWithColorTag(c color.Attribute, msg string) {
	color.Set(c)
	fmt.Printf(TAG + msg + "\n")
	color.Unset()
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
