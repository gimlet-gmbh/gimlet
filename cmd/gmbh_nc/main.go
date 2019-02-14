package main

import (
	"flag"
	"os"

	"github.com/gmbh-micro/notify"
)

func main() {

	configPath := flag.String("config", "", "the path to the gmbh config file (yaml)")
	flag.Parse()

	if *configPath == "" {
		notify.LnRedF("must specify config file")
		os.Exit(1)
	}

	notify.StdMsgBlue("newCore")
	c, err := NewCore(*configPath)
	if err != nil {
		panic(err)
	}
	c.Start()
}
