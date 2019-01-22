package gprint

import (
	"fmt"

	"github.com/fatih/color"
)

/**
 * gprint.go
 * Abe Dick
 * January 2019
 */

// TAG - the heading of the line
const TAG = "[gmbh] "

// Ln prints a single line w/ tag
func Ln(msg string, indent int) {
	color.Set(color.FgBlue)
	out(setIndent(indent) + msg)
	color.Unset()
}

// Cln prints a custom single line w/ tag
func Cln(tag, msg string, indent int, ccolor color.Attribute) {
	color.Set(ccolor)
	fmt.Println(tag + setIndent(indent) + msg)
	color.Unset()
}

// Err prints msg in red
func Err(msg string, indent int) {
	color.Set(color.FgRed)
	out(setIndent(indent) + msg)
	color.Unset()
}

// Green prints msg in green
func Green(msg string, indent int) {
	color.Set(color.FgGreen)
	out(setIndent(indent) + msg)
	color.Unset()
}

// Log prints msg in Cyan
func Log(msg string, indent int) {
	color.Set(color.FgCyan)
	out(setIndent(indent) + msg)
	color.Unset()
}

// Update msg in Cyan
func Update(msg string, indent int) {
	color.Set(color.FgCyan)
	out(setIndent(indent) + msg)
	color.Unset()
}

func setIndent(indent int) string {
	r := ""
	for i := 0; i < indent; i++ {
		r += "    "
	}
	return r
}

func out(msg string) {
	fmt.Println(TAG + msg)
}
