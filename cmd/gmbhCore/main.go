package main

import (
	"flag"
	"os"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
)

func main() {

	configPath := flag.String("config", "", "the path to the gmbh config file (toml)")
	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")
	flag.Parse()

	coreAddr := config.DefaultSystemCore.Address
	env := os.Getenv("ENV")
	if env == "C" {
		coreAddr = os.Getenv("CORE")
	}

	c, err := NewCore(*configPath, env, coreAddr, *verbose)
	if err != nil {
		panic(err)
	}
	c.Start()
}

func print(format string, a ...interface{}) {
	if core.env == "M" {
		format = "[" + time.Now().Format(config.LogStamp) + "] [core] " + format
		notify.LnCyanF(format, a...)
	} else {
		notify.LnCyanF("[core] "+format, a...)
	}
}
