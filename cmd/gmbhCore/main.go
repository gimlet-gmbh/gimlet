package main

import (
	"flag"
	"os"

	"github.com/gmbh-micro/notify"
)

func main() {

	configPath := flag.String("config", "", "the path to the gmbh config file (yaml)")
	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")
	verbosedata := flag.Bool("verbose-data", false, "print gmbh data output to stdOut and stdErr")

	/*
		Things that should be paremetrized in Core
		* Should services be in their own remote or all in one remote?
	*/

	flag.Parse()

	if *configPath == "" {
		notify.LnRedF("must specify config file")
		os.Exit(1)
	}
	c, err := NewCore(*configPath, *verbose, *verbosedata)
	if err != nil {
		panic(err)
	}
	c.Start()
}
