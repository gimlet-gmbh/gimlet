package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

func main() {

	// modes
	remote := flag.Bool("remote", false, "begin a gmbhRemote service; must specify a config file")
	core := flag.Bool("core", false, "begin a gmbhCore instance; must specify a config file")

	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")
	verbosedata := flag.Bool("verbose-data", false, "print gmbhData output to stdOut and stdErr")
	nolog := flag.Bool("no-log", false, "disable logging")
	daemon := flag.Bool("daemon", false, "for running the process manager in a container")

	config := flag.String("config", "", "a gmbh configuration file path")

	listAllFlag := flag.Bool("list", false, "list all processes")
	reportFlag := flag.Bool("report", false, "list all processes in report format")
	restartAllFlag := flag.Bool("restart", false, "restart all processes")
	listOneFlag := flag.String("list-one", "", "list all processes")
	restartOneFlag := flag.String("restart-one", "", "list all processes")
	shutdownFlag := flag.Bool("q", false, "shutdown gmbh")

	flag.Parse()

	if *remote {
		startRemote(*config)
	} else if *core {
		startCore(*config, *verbose, *verbosedata, *daemon, *nolog)

	} else if *listAllFlag {
		listAll()
	} else if *reportFlag {
		runReport()
	} else if *restartAllFlag {
		restartAll()
	} else if *listOneFlag != "" {
		listOne(*listOneFlag)
	} else if *restartOneFlag != "" {
		restartOne(*restartOneFlag)
	} else if *shutdownFlag {
		shutdown()
	} else {
		notify.LnRedF("rerun with --help to see options")
	}

}

func startRemote(c string) {

}

func startCore(c string, verbose, vdata, daemon, nolog bool) {
	report()

	installed := checkInstall()
	if !installed {
		notify.LnBRedF("gmbhCore does not seem to be installed")
		os.Exit(1)
	}

	if c == "" {
		notify.LnBRedF("must specify a config file using the --config flag")
		os.Exit(1)
	}

	exists := checkConfig(c)
	if !exists {
		notify.LnBRedF("the specified config file does not seem to exist...")
		os.Exit(1)
	}

	var pmlog *os.File
	var datalog *os.File
	var err error

	pmCmd := exec.Command("gmbhProcm")
	gmbhCmd := exec.Command("gmbhCore", "--config="+c)

	if verbose {
		pmCmd.Stdout = os.Stdout
		pmCmd.Stderr = os.Stderr
		pmCmd.Args = append(pmCmd.Args, "--verbose")

		gmbhCmd.Stdout = os.Stdout
		gmbhCmd.Stderr = os.Stderr
		gmbhCmd.Args = append(gmbhCmd.Args, "--verbose")
	}
	if vdata {
		pmlog, err = getLogFile("gmbh", "procm.log")
		if err == nil {
			notify.LnYellowF("logs")
			notify.LnYellowF(filepath.Join(getpwd(), "gmbh", "procm.log"))
			pmCmd.Stdout = pmlog
			pmCmd.Stderr = pmlog

			gmbhCmd.Stdout = os.Stdout
			gmbhCmd.Stderr = os.Stderr
			gmbhCmd.Args = append(gmbhCmd.Args, "--verbose-data")
		}
	}
	if !verbose && !vdata && !nolog {
		notify.LnYellowF("logs")
		pmlog, err = getLogFile("gmbh", "procm.log")
		if err == nil {
			notify.LnYellowF(filepath.Join(getpwd(), "gmbh", "procm.log"))
			pmCmd.Stdout = pmlog
			pmCmd.Stderr = pmlog
		} else {
			panic(err)
		}
		datalog, err = getLogFile("gmbh", "data.log")
		if err == nil {
			notify.LnYellowF(filepath.Join(getpwd(), "gmbh", "data.log"))
			gmbhCmd.Stdout = datalog
			gmbhCmd.Stderr = datalog
		} else {
			panic(err)
		}
	}

	remoteEnv := append(
		os.Environ(),
		"PMMODE=PMManaged",
	)
	pmCmd.Env = remoteEnv
	gmbhCmd.Env = remoteEnv

	err = pmCmd.Start()
	if err != nil {
		notify.LnBRedF("could not start gmbh-procm")
		return
	}
	err = gmbhCmd.Start()
	if err != nil {
		notify.LnBRedF("could not start gmbh-data")
		return
	}

	if !daemon {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT)

		notify.LnBBlueF("holding until shutdown signal")
		_ = <-sig
		fmt.Println() //dead line to line up output

		// signal the processes
		notify.LnBBlueF("signaled sigusr1")
		pmCmd.Process.Signal(syscall.SIGUSR1)

		// shutdown gmbh
		time.Sleep(time.Second * 1)
		gmbhCmd.Process.Signal(syscall.SIGUSR2)
		gmbhCmd.Wait()
		notify.LnBYellowF("[cli] gmbh shutdown")

		// shutdown the process manager
		time.Sleep(time.Second * 1)
		pmCmd.Process.Signal(syscall.SIGUSR2)
		pmCmd.Wait()
		notify.LnBYellowF("[cli] procm shutdown")

		// close the logs
		if pmlog != nil {
			pmlog.Close()
		}
		if datalog != nil {
			datalog.Close()
		}

		notify.LnBYellowF("[cli] shutdown complete")
	}
}

// getLogFile attempts to add the desired path as an extension to the current
// directory as reported by os.GetWd(). The file is then opened or created
// and returned
func getLogFile(desiredPathExt, filename string) (*os.File, error) {
	// get pwd
	dir, err := os.Getwd()
	if err != nil {
		notify.LnBRedF("getlogfile, pwd err=%s", err.Error())
		return nil, err
	}
	// make sure that the path extension exists or make the directories needed
	dirPath := filepath.Join(dir, desiredPathExt)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
	}
	// create the file
	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		notify.LnBRedF("openfile err=%s", err.Error())
		return nil, err
	}
	return file, nil
}

func getpwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func checkConfig(configPath string) bool {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkInstall() bool {
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/gmbhCore"); os.IsNotExist(err) {
			return false
		}
		return true
	} else if runtime.GOOS == "linux" {
		notify.StdMsgErr("Linux support is incomplete")
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/gmbhCore"); os.IsNotExist(err) {
			return false
		}
		return true
	}
	notify.StdMsgErr(fmt.Sprintf("OS support not yet implemented for %s", runtime.GOOS))
	return false
}

func report() {
	notify.LnBCyanF("                   ")
	notify.LnBCyanF("  _  ._ _  |_  |_| ")
	notify.LnBCyanF(" (_| | | | |_) | | ")
	notify.LnBCyanF("  _|               ")
	notify.LnBCyanF("Version=%s; Code=%s", defaults.VERSION, defaults.CODE)
}
