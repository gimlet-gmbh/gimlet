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
	"github.com/gmbh-micro/notify"
)

func init() {
	if len(os.Args) < 2 {
		notify.StdMsgErr("must start gmbhCore with argument containing path to project")
		os.Exit(1)
	}
}

func main() {
	daemon := false
	if len(os.Args) == 3 {
		if os.Args[2] == "-d" {
			daemon = true
		}
	}

	gmbhCore := StartCore(os.Args[1], true, daemon)

	printLogo()
	notify.StdMsgBlue("Starting version: " + gmbhCore.Version + " (" + gmbhCore.CodeName + ")")

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
