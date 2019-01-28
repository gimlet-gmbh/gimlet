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
	"github.com/gmbh-micro/core"
	"github.com/gmbh-micro/notify"
)

func init() {
	if len(os.Args) < 2 {
		notify.StdMsgErr("must start gmbhCore with argument containing path to project")
		os.Exit(1)
	}
}

func main() {

	gmbhCore := core.StartCore(os.Args[1])

	printLogo()
	notify.StdMsgBlue("Starting version: "+gmbhCore.Version+" ("+gmbhCore.CodeName+")", 0)

	gmbhCore.StartCabalServer()
	gmbhCore.StartControlServer()
	gmbhCore.ServiceDiscovery()
}

func printLogo() {
	color.Set(color.FgBlue)
	fmt.Println("                    _           ")
	fmt.Println("  _  ._ _  |_  |_| /   _  ._ _  ")
	fmt.Println(" (_| | | | |_) | | \\_ (_) | (/_")
	fmt.Println("  _|                            ")
	color.Unset()
}
