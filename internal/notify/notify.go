package notify

import (
	"fmt"

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

// LnF prints to stdOut a line formatted
func LnF(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
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
