package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
)

func main() {

	// modes
	// remote := flag.Bool("remote", false, "begin a gmbhRemote service; must specify a config file")
	core := flag.Bool("core", false, "begin a gmbhCore instance; must specify a config file")
	noServiceDiscovery := flag.Bool("noSDiscovery", false, "do not start managed processes")

	verbosedata := flag.Bool("verbose", false, "print gmbhCoreData output to stdOut and stdErr")
	verbose := flag.Bool("verbose-data", false, "print all output to stdOut and stdErr")
	nolog := flag.Bool("no-log", false, "disable logging")
	daemon := flag.Bool("daemon", false, "for running the process manager in a container")

	configs := flag.String("config", "", "a gmbh configuration file path")

	listAllFlag := flag.Bool("list", false, "list all processes")
	reportFlag := flag.Bool("report", false, "list all processes in report format")
	restartAllFlag := flag.Bool("restart", false, "restart all processes")
	listOneFlag := flag.String("list-one", "", "list all processes")
	restartOneFlag := flag.String("restart-one", "", "list all processes")
	shutdownFlag := flag.Bool("q", false, "shutdown gmbh")

	flag.Parse()

	setCore(*configs)

	if *core {
		startCore(*configs, *verbose, *verbosedata, *daemon, *nolog, *noServiceDiscovery)
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
		startCore(*configs, *verbose, *verbosedata, *daemon, *nolog, *noServiceDiscovery)
	}

}

func startServiceDiscovery(cfile string, verbose, daemon bool) {
	userConfig, err := config.ParseSystemConfig(cfile)
	if err != nil {
		notify.LnBRedF("could not parse config; err=%s", err.Error())
		return
	}
	path := filepath.Join(basePath(cfile), userConfig.Services.ServicesDirectory)
	servicePaths, err := scanForServices(path)
	if err != nil {
		notify.LnBRedF("error scanning for services; err=%s", err.Error())
		return
	}

	// Create and attach all services that run in Managed mode
	for _, servicePath := range servicePaths {

		static, err := config.ParseServiceStatic(servicePath)
		if err != nil {
			notify.LnBRedF("could not open config file")
			continue
		}

		if ok := static.Validate(); ok != nil {
			notify.LnBRedF("could not validate config file")
			continue
		}
		launchService(servicePath, servicePath, config.DefaultSystemCore.Address, verbose)
	}

}

func setCore(configPath string) {

	Service := config.ServiceConfig{
		Static: &config.ServiceStatic{
			Name:     "core",
			Language: "go",
			BinPath:  "gmbhCore",
			Args:     []string{"--config=" + configPath, "--verbose"},
		},
	}

	dir := filepath.Dir(configPath)
	coreServiceConfig := filepath.Join(dir, ".core.service")

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(Service); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(coreServiceConfig)
	if err != nil {
		log.Fatal(err)
	}
	w := bufio.NewWriter(f)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
	f.Close()

}

func startCore(c string, verbose, vdata, daemon, nolog, noServiceDiscovery bool) {
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

	gmbhCmd := exec.Command("gmbhProcm", "--remote", "--config=./.core.service", "--verbose")
	// gmbhCmd.Stdout = os.Stdout
	// gmbhCmd.Stderr = os.Stderr
	workingEnv := []string{
		"SERVICEMODE=managed",
	}

	if verbose {
		pmCmd.Stdout = os.Stdout
		pmCmd.Stderr = os.Stderr
		pmCmd.Args = append(pmCmd.Args, "--verbose")

		gmbhCmd.Stdout = os.Stdout
		gmbhCmd.Stderr = os.Stderr
		workingEnv = append(
			workingEnv,
			"REMOTELOG="+filepath.Join(basePath(c), "gmbh", "data-remote.log"),
		)
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
			gmbhCmd.Args = append(gmbhCmd.Args, "--verbose")
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
		}
		workingEnv = append(
			workingEnv,
			"REMOTELOG="+filepath.Join(basePath(c), "gmbh", "data-remote.log"),
		)

	}

	remoteEnv := append(
		os.Environ(),
		"SERVICEMODE=managed",
	)
	pmCmd.Env = remoteEnv

	gmbhCmd.Env = append(os.Environ(), workingEnv...)

	err = pmCmd.Start()
	if err != nil {
		notify.LnBRedF("could not start gmbh-procm")
		return
	}
	err = gmbhCmd.Start()
	if err != nil {
		notify.LnBRedF("could not start gmbh-core-data")
		return
	}

	if !noServiceDiscovery {
		go startServiceDiscovery(c, verbose, daemon)
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

		// shutdown the process manager
		time.Sleep(time.Second * 3)
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
		notify.LnRedF("Linux support is incomplete")
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/gmbhCore"); os.IsNotExist(err) {
			return false
		}
		return true
	}
	notify.LnRedF(fmt.Sprintf("OS support not yet implemented for %s", runtime.GOOS))
	return false
}

func report() {
	// notify.LnBCyanF("                   ")
	// notify.LnBCyanF("  _  ._ _  |_  |_| ")
	// notify.LnBCyanF(" (_| | | | |_) | | ")
	// notify.LnBCyanF("  _|               ")
	notify.LnBCyanF("                    __                                                ")
	notify.LnBCyanF("  _  ._ _  |_  |_  (_   _  ._   o  _  _  |   _.     ._   _ |_   _  ._ ")
	notify.LnBCyanF(" (_| | | | |_) | | __) (/_ | \\/ | (_ (/_ |_ (_| |_| | | (_ | | (/_ |  ")
	notify.LnBCyanF("  _|                                                                  ")
	notify.LnBCyanF("Version=%s; Code=%s", config.Version, config.Code)
}
