package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

func main() {

	// modes
	// cli := flag.Bool("i", false, "retrieve information from gmbH when running")
	remote := flag.Bool("remote", false, "begin a gmbhRemote service; must specify a config file")
	core := flag.Bool("core", false, "begin a gmbhCore instance; must specify a config file")

	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")
	verbosedata := flag.Bool("verbose-data", false, "print gmbhData output to stdOut and stdErr")

	// daemon := flag.Bool("daemon", false, "print all output to a log file")

	config := flag.String("config", "", "a gmbh configuration file path")

	// listAllFlag := flag.Bool("list", false, "list all processes")
	// reportFlag := flag.Bool("report", false, "list all processes in report format")
	// restartAllFlag := flag.Bool("restart", false, "restart all processes")
	// listOneFlag := flag.String("list-one", "", "list all processes")
	// restartOneFlag := flag.String("restart-one", "", "list all processes")
	// shutdownFlag := flag.Bool("q", false, "shutdown gmbh")

	flag.Parse()
	// if *mode {

	// }

	switch {
	case *remote:
		startRemote(*config)
	case *core:
		startCore(*config, *verbose, *verbosedata)

	default:
		notify.LnRedF("rerun with --help to see options")
	}

}

func startRemote(c string) {

}

func startCore(c string, v, vd bool) {
	report()

	pmcmd := exec.Command("gmbhPM")
	pmcmd.Env = os.Environ()

	if v {
		pmcmd.Stdout = os.Stdout
		pmcmd.Stderr = os.Stderr
		pmcmd.Args = append(pmcmd.Args, "--verbose")
	}

	err := pmcmd.Start()
	if err != nil {
		notify.LnRedF("could not start remote")
		return
	}

	gmbHcmd := exec.Command("gmbhNC", "--config="+c)
	gmbHcmd.Env = os.Environ()

	if v || vd {
		gmbHcmd.Stdout = os.Stdout
		gmbHcmd.Stderr = os.Stderr
		if v {
			gmbHcmd.Args = append(gmbHcmd.Args, "--verbose")
		} else {
			gmbHcmd.Args = append(gmbHcmd.Args, "--verbose-data")
		}
	}

	err = gmbHcmd.Start()
	if err != nil {
		notify.LnRedF("could not start remote")
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	notify.LnBGreenF("main thread waiting")
	_ = <-sig
	fmt.Println() //dead line to line up output
}

func startCLI() {

}

func report() {
	notify.LnBCyanF("                   ")
	notify.LnBCyanF("  _  ._ _  |_  |_| ")
	notify.LnBCyanF(" (_| | | | |_) | | ")
	notify.LnBCyanF("  _|               ")
	notify.LnBCyanF("Version=%s; Code=%s", defaults.VERSION, defaults.CODE)
}
