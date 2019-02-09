package main

/*
 * main.go
 * Abe Dick
 * January 2019
 */

import (
	"flag"
	"os"
	"time"

	"github.com/gmbh-micro/notify"
)

func main() {

	path := flag.String("path", "", "the path to the directory of the gmbh Config")
	// container := flag.String("c", false, "")
	flag.Parse()

	if *path == "" {
		notify.StdMsgErr("cannot start gmbhCore without path argument")
		os.Exit(1)
	}

	core, err := StartCore(*path)
	if err != nil {
		core.Log.Err("could not start gmbhCore; error=%v", err.Error())
		return
	}

	notify.StdMsgBlueNoPrompt("                    _           ")
	notify.StdMsgBlueNoPrompt("  _  ._ _  |_  |_| /   _  ._ _  ")
	notify.StdMsgBlueNoPrompt(" (_| | | | |_) | | \\_ (_) | (/_")
	notify.StdMsgBlueNoPrompt("  _|                            ")
	notify.StdMsgBlue("version=" + core.Version + "; code=" + core.CodeName + "; startTime=" + core.StartTime.Format(time.Stamp))

	err = core.StartCabalServer()
	if err != nil {
		notify.StdMsgErr("could not start cabal server")
		return
	}

	err = core.StartControlServer()
	if err != nil {
		notify.StdMsgErr("could not start control server")
		return
	}

	err = core.ServiceDiscovery()
	if err != nil {
		notify.StdMsgErr("service discovery error")
	}

	notify.StdMsgBlueNoPrompt(notify.SEP)

	core.Wait()
}
