package main

/**
 * main.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/gimlet-gmbh/gimlet/gprint"
)

const (
	VERSION    = "00.02.00"
	PROMPT     = "[cli] "
	CONFIGNAME = "gimlet.yaml"
	GIMLETPATH = "/usr/local/bin/gimlet"
)

var cmd *exec.Cmd

func main() {
	startGimlet()
}

func startGimlet() {
	println(fmt.Sprintf("cli version: %s", VERSION), 0)
	println("Starting Gimlet...", 0)

	exists := checkConfig()
	if !exists {
		printerr("could not find config file", 1)
		return
	}
	println("found config", 1)

	exists = checkInstall()
	if !exists {
		printerr("could not find gimlet", 1)
		return
	}
	println("found gimlet binary", 1)

	// Monitor/ Force Core Shutdown
	wg := new(sync.WaitGroup)
	wg.Add(1)
	shutdownSignal := make(chan os.Signal, 1)

	go startListener(shutdownSignal, wg)

	// Fork/Exec gimlet
	forkExec(GIMLETPATH, []string{getCurrentDir()})

}

func checkConfig() bool {
	if _, err := os.Stat("gimlet.yaml"); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkInstall() bool {
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(GIMLETPATH); os.IsNotExist(err) {
			return false
		}
		return true
	}
	println(fmt.Sprintf("OS support not yet implemented for %s", runtime.GOOS), 0)
	return false
}

func startListener(sig chan os.Signal, wg *sync.WaitGroup) {

	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGKILL)

	_ = <-sig

	fmt.Println("")
	time.Sleep(time.Millisecond * 500)

	// Kill the process
	kill()

	wg.Done()
	os.Exit(0)
}

func forkExec(path string, args []string) {
	cmd = setCmd(path, args)

	err := cmd.Run()
	if err != nil {
		printerr(fmt.Sprintf("Error reported in Core: %s", err.Error()), 0)
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
	gprint.Cln(PROMPT, msg, indent, color.FgHiMagenta)
}

func printerr(msg string, indent int) {
	gprint.Cln(PROMPT, msg, indent, color.FgRed)
}

// func genConfig() {

// }
