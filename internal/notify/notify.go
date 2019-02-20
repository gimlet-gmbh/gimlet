package notify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// SEP erator between things
const SEP = "-----------------------------------------------"

var verbose = true
var header = ""

// SetVerbose on or off
func SetVerbose(on bool) {
	verbose = on
}

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
**** OS Helpers (These should probably migrate somewhere else)
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
