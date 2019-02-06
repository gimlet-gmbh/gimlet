package main

/*
 * main.go (gmbhContainer)
 * Abe Dick
 * February 2019
 */

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/service"
)

const PATH_TO_CONFIG = "/config/gmbh.yaml"

// First arg should be path to file where gmbh-config can be found

func main() {
	notify.SetTag("[gmbh-exp] ")
	notify.StdMsgBlue("gmbh container process manager")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		notify.StdMsgErr("could not create absolute file path")
	}

	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		notify.StdMsgErr("no path specified")
		os.Exit(1)
	}
	notify.StdMsgBlue("Path: " + path)

	s, err := service.NewService(path, service.Remote)
	s.StartLog(dir, "process-manager.log")
	if err != nil {
		notify.StdMsgErr("could not create service: " + err.Error())
		os.Exit(1)
	}
	notify.StdMsgBlue("Attempting to start " + s.Static.Name)
	pid, err := s.StartService()
	if err != nil {
		notify.StdMsgErr("could not start service: " + err.Error())
		os.Exit(1)
	}
	notify.StdMsgGreen("service process started with pid=" + pid)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()

	notify.StdMsgDebug("Blocking until control server interface is implemented")
	<-done
	s.KillProcess()

	fmt.Println("")
	notify.StdMsgGreen("Shutdown signal received")

	os.Exit(0)
}
