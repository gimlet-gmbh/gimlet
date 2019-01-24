package main

/*
 * main.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gimlet-gmbh/gimlet/gcore"
	"github.com/gimlet-gmbh/gimlet/notify"
)

func init() {
	if len(os.Args) < 2 {
		notify.StdMsgErr("must start gimlet with argument containing path to project")
		os.Exit(1)
	}
}

func main() {

	core := gcore.StartCore(os.Args[1])

	printLogo()
	notify.StdMsgBlue("Starting version: "+core.Version+" ("+core.CodeName+")", 0)

	core.StartInternalServer()
	core.ServiceDiscovery()
}

func printLogo() {
	color.Set(color.FgBlue)
	fmt.Println("                    _           ")
	fmt.Println("  _  ._ _  |_  |_| /   _  ._ _  ")
	fmt.Println(" (_| | | | |_) | | \\_ (_) | (/_")
	fmt.Println("  _|                            ")
	color.Unset()
}
