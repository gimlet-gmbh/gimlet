package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"flag"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

const (
	DEBUG = true
)

func main() {
	notify.SetTag(defaults.CTRL_PROMPT)
	if DEBUG {
		notify.StdMsgBlue("gmbhCtrl Tool")
		notify.StdMsgBlue("starting version " + defaults.VERSION)
		notify.StdMsgBlue("running in debug mode")
	}

	listAllFlag := flag.Bool("list", false, "list all processes")
	reportFlag := flag.Bool("report", false, "list all processes in report format")
	restartAllFlag := flag.Bool("restart", false, "restart all processes")
	listOneFlag := flag.String("list-one", "", "list all processes")
	restartOneFlag := flag.String("restart-one", "", "list all processes")
	shutdownFlag := flag.Bool("q", false, "shutdown gmbh")
	flag.Parse()

	if *listAllFlag {
		listAll()
	} else if *reportFlag {
		report()
	} else if *restartAllFlag {
		restartAll()
	} else if *listOneFlag != "" {
		listOne(*listOneFlag)
	} else if *restartOneFlag != "" {
		restartOne(*restartOneFlag)
	} else if *shutdownFlag {
		shutdown()
	}

}
