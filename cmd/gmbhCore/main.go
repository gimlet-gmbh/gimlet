package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gmbh-micro/notify"
)

func main() {

	configPath := flag.String("config", "", "the path to the gmbh config file (toml)")
	address := flag.String("address", "", "specifying an address here can be used in place of a config (All defaults will be used except for the address); note that if a config file is also specified, this is the address that will be used")
	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")

	flag.Parse()

	c, err := NewCore(*configPath, *address, *verbose)
	if err != nil {
		panic(err)
	}
	c.Start()
}

func logData(msg string, a ...interface{}) {
	notify.LnCyanF(fmt.Sprintf("[%s] %s", time.Now().Format(time.Stamp), msg), a...)
}
func logCore(msg string, a ...interface{}) { notify.LnCyanF("[core] "+msg, a...) }
func logRtr(msg string, a ...interface{})  { notify.LnCyanF("[rtr] "+msg, a...) }
