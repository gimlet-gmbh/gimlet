package main

/**
 * main.go
 * Abe Dick
 * January 2019
 */

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

///////////////////////////////////
// This is garbage. Use the actual process manager...

var cmd *exec.Cmd

func main() {

	notify.SetTag(defaults.CLI_PROMPT)
	notify.SetVerbose(defaults.VERBOSE)
	daemon := flag.Bool("d", defaults.DAEMON, "daemon mode")
	flag.Parse()
	startCore(*daemon)
}

func startCore(daemon bool) {
	println(fmt.Sprintf("cli version: %s", defaults.VERSION), 0)
	println("Starting gmbhCore...", 0)

	exists := checkConfig()
	if !exists {
		printerr("could not find config file", 1)
		return
	}
	println("found config", 1)

	exists = checkInstall()
	if !exists {
		printerr("could not find gmbhCore", 1)
		return
	}
	println("found gmbhCore binary", 1)

	// Monitor/ Force Core Shutdown
	wg := new(sync.WaitGroup)
	wg.Add(1)
	shutdownSignal := make(chan os.Signal, 1)

	go startListener(shutdownSignal, wg)

	// Fork/Exec gmbhCore
	args := []string{getCurrentDir()}
	if daemon {
		args = append(args, "-d")
	}
	forkExec(defaults.CORE_PATH_MAC, args, daemon)

}

func checkConfig() bool {
	if _, err := os.Stat(defaults.PROJECT_CONFIG_FILE); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkInstall() bool {
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(defaults.CORE_PATH_MAC); os.IsNotExist(err) {
			return false
		}
		return true
	} else if runtime.GOOS == "linux" {
		notify.StdMsgBlue("Linux support is not complete")
		if _, err := os.Stat(defaults.CORE_PATH_MAC); os.IsNotExist(err) {
			return false
		}
		return true
	}
	notify.StdMsgErr(fmt.Sprintf("OS support not yet implemented for %s", runtime.GOOS))
	return false
}

func startListener(sig chan os.Signal, wg *sync.WaitGroup) {

	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGKILL)

	_ = <-sig

	fmt.Println("")
	time.Sleep(time.Millisecond * 500)

	kill()

	wg.Done()
	os.Exit(0)
}

func forkExec(path string, args []string, daemon bool) {
	cmd = setCmd(path, args)

	if daemon {
		err := cmd.Start()
		if err != nil {
			notify.StdMsgErr(fmt.Sprintf("Error reported in Core: %s", err.Error()))
		}
	} else {
		err := cmd.Run()
		if err != nil {
			notify.StdMsgErr(fmt.Sprintf("Error reported in Core: %s", err.Error()))
		}
	}
}

func kill() {
	if cmd != nil {
		err := cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			printerr(err.Error(), 0)
		}
	} else {
		printerr("cmd was never set", 0)
	}
}

func setCmd(path string, args []string) *exec.Cmd {
	var cmd *exec.Cmd
	cmd = exec.Command(path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func getCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		printerr(fmt.Sprintf("could not get current dir: %s", err.Error()), 0)
	}
	return dir
}

func println(msg string, indent int) {
	notify.StdMsgMagenta(msg, indent)
}

func printerr(msg string, indent int) {
	notify.StdMsgErr(msg, indent)
}

// func genConfig() {

// }
