package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"github.com/gmbh-micro/notify"
)

const (
	VERSION    = "00.07.01"
	CTRLSERVER = "localhost:59997"
)

func main() {
	notify.SetTag("[ctrl] ")
	notify.StdMsgBlue("gmbhCtrl Tool")
	notify.StdMsgBlue("starting version " + VERSION)

}
