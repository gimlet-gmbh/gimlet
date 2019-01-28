package main

import "github.com/gmbh-micro/notify"

const (
	VERSION = "00.07.00"
)

func main() {
	notify.SetTag("[ctrl] ")
	notify.StdMsgBlue("starting version " + VERSION)
}
