package gflags

/**
 * gflags.go
 * Abe Dick
 * December 2018
 */

import (
	"fmt"
	"os"
	"strings"
)

var formals *flagSet

// Parse command line arguments, panics on error
func Parse() {
	if formals != nil {
		formals.parse()
	}
}

func parseFormals() *flagSet {
	return &flagSet{}
}

func checkSet() {
	if formals == nil {
		initSet()
	}
}

func initSet() {
	formals = parseFormals()
	formals.word = make(map[string]*Flag)
	formals.letter = make(map[string]*Flag)
}

// flagSet stores the location of all of the possible parsed flags
type flagSet struct {
	word   map[string]*Flag
	letter map[string]*Flag
}

// A Flag represents the state of a "formal" flag.
// Instead of using the value interface method as in the golang flag pkg, all
// flags will have an internal boolean representation (if 0 args, boolValue=true)
// else false. The only other type of value is
type Flag struct {
	Name       string  // name as it appears on command line if word
	LetterName string  // name as it appears on command line if letter
	Usage      string  // help message
	StringArgs string  // help message for string usage
	BoolValue  bool    // true if boolean flag else false
	retBool    *bool   // Will have an address if wanting to return a boolean from flag
	retStr     *string // Will have an address if wanting to return a string form flag
	Parsed     bool    // if the flag has been parsed
}

// An intermediate value for the flag to use in between gathering cmd line args and
// getting the actual to set the value with
type actual struct {
	name   string
	letter bool
	args   []string
}

// SetBool registeres boolean command line flags
func SetBool(word, letter, usage, strUsage string, ptr *bool) {
	registerFlag(&Flag{
		Name:       word,
		LetterName: letter,
		Usage:      usage,
		StringArgs: strUsage,
		BoolValue:  true,
		retBool:    ptr,
	})
}

// SetString registers string command line flags
func SetString(word, letter, usage, strUsage string, ptr *string) {
	registerFlag(&Flag{
		Name:       word,
		LetterName: letter,
		Usage:      usage,
		StringArgs: strUsage,
		BoolValue:  false,
		retStr:     ptr,
	})
}

// registeres a flag
func registerFlag(newFlag *Flag) {
	checkSet()
	formals.registerFlag(newFlag)
}

func (f *flagSet) registerFlag(newFlag *Flag) {
	if _, exists := formals.word[newFlag.Name]; exists {
		panic("flag name collision")
	}
	if _, exists := formals.letter[newFlag.LetterName]; exists {
		panic("flag letter collision")
	}
	formals.word[newFlag.Name] = newFlag
	formals.letter[newFlag.LetterName] = newFlag
}

func (f *flagSet) parse() {
	args := os.Args[1:]
	iFlags := []actual{}
	i, j := 0, 0
	for i < len(args) {
		if args[i][0] == '-' {
			for j < len(args)-1 {
				j++
				if args[j][0] == '-' {
					j--
					break
				}
			}
		}
		iFlags = append(iFlags, f.processFlag(args[i:j+1]))
		j++
		i = j
	}
	f.setFlags(iFlags)
	for _, flag := range f.letter {
		if !flag.Parsed {
			if flag.BoolValue {
				*flag.retBool = false
			} else {
				*flag.retStr = ""
			}
			flag.Parsed = true
		}
	}
}

func (f *flagSet) setFlags(iflags []actual) {
	for _, flag := range iflags {
		var actual *Flag
		if flag.letter {
			actual = f.letter[flag.name]
		} else {
			actual = f.word[flag.name]
		}

		if actual == nil {
			f.printError()
			os.Exit(1)
		} else {
			if actual.BoolValue {
				*actual.retBool = true
			} else {
				*actual.retStr = strings.Join(flag.args, " ")
			}
		}
		actual.Parsed = true
	}
}

func (f *flagSet) processFlag(arg []string) actual {
	if arg == nil || len(arg) == 0 {
		return actual{}
	}
	ret := actual{letter: false}
	if len(arg[0]) > 1 {
		if arg[0][0:2] == "--" {
			arg[0] = arg[0][2:]
		} else if arg[0][0] == '-' {
			arg[0] = arg[0][1:]
			ret.letter = true
		}
	}
	if arg[0] == "h" || arg[0] == "help" {
		f.usage()
		os.Exit(0)
	}
	ret.name = arg[0]
	ret.args = arg[1:]
	return ret
}

func (f *flagSet) printError() {
	fmt.Println("Error parsing flags.")
	f.usage()
}

func (f *flagSet) usage() {
	fmt.Printf("\nUsage:\n")
	for _, v := range f.word {
		fmt.Printf("  -%s --%s %s\n\t%s\n\n", v.LetterName, v.Name, v.StringArgs, v.Usage)
	}
}
