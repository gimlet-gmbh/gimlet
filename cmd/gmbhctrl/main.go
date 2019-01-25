package main

import "github.com/gimlet-gmbh/gimlet/notify"

const (
	VERSION = "00.00.01"
)

func main() {
	notify.SetTag("[ctrl] ")
	notify.StdMsgBlue("starting version " + VERSION)
}
