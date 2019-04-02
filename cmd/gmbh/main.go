package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/fileutil"
	"github.com/gmbh-micro/notify"
	"github.com/rs/xid"
)

const (
	coreService = "core.toml"
)

var (
	procmcmd *exec.Cmd
	procmlog *os.File

	corecmd *exec.Cmd
	corelog *os.File

	fingerprint = xid.New().String()
)

func main() {

	run := flag.Bool("run", false, "runs a local, managed gmbh server instance. Must specify the config file with the --config flag")
	deploy := flag.Bool("build-deploy", false, "build a gmbh managed docker cluster")
	report := flag.Bool("info", false, "combine info with any of the ")

	config := flag.String("config", "", "a gmbh configuration file path")
	daemon := flag.Bool("daemon", false, "gmbh will run without printing to std[out,err,log]")
	verbose := flag.Bool("verbose", false, "Prints all Core data to stdout and stderr; rest goes to log files. Overrides config file verbose setting")
	nolog := flag.Bool("no-log", false, "disable logging")

	listall := flag.Bool("list", false, "list all processes")
	listone := flag.String("list-one", "", "list all processes")
	restartall := flag.Bool("restart", false, "restart all processes")
	restartone := flag.String("restart-one", "", "list all processes")
	q := flag.Bool("q", false, "shutdown gmbh")

	flag.Parse()

	if *run || *deploy {

		if *config == "" {
			notify.LnRedF("error, must specify a gmbh config file")
			os.Exit(1)
		}

		if *run {
			start(*config, *verbose, *nolog, *daemon)
		} else {
			builddeploy(*config)
		}

	} else if *report {
		if *listall {
			listAll()
		} else if *listone != "" {
			listOne(*listone)
		} else if *restartall {
			restartAll()
		} else if *restartone != "" {
			restartOne(*restartone)
		} else if *q {
			shutdown()
		}
	}
}

// start validates if gmbh is installed and checks if the core service file has been created
func start(cfile string, verbose, nolog, daemon bool) {

	notify.LnF("                  ")
	notify.LnF("  _  ._ _  |_  |_|")
	notify.LnF(" (_| | | | |_) | |")
	notify.LnF("  _|              ")
	notify.LnF("Version=%s; Code=%s", config.Version, config.Code)

	// make sure that gmbhCore is installed
	installed := checkInstall()
	if !installed {
		notify.LnRedF("gmbhCore or gmbhProcm does not seem to be installed")
		os.Exit(1)
	}

	conf, err := config.ParseSystemConfig(cfile)
	if err != nil {
		notify.LnRedF("specified config file cannot be parsed, err=%s", err.Error())
		os.Exit(1)
	}

	fileutil.MkDir("gmbh")

	if !fileExists(filepath.Join("gmbh", coreService)) {
		notify.LnF("Generating core service config file...")
		err = genCoreConf(filepath.Join("gmbh", coreService), cfile, conf)
		if err != nil {
			notify.LnRedF("cannot create core config file, err=%s", err.Error())
			os.Exit(1)
		}
	}

	err = startProcm(nolog)
	if err != nil {
		notify.LnRedF("could not start ProcM, err=%s", err.Error())
		os.Exit(1)
	}

	err = startCore(nolog, verbose)
	if err != nil {
		notify.LnRedF("could not start core, err=%s", err.Error())
		os.Exit(1)
	}

	startServices(conf)

	if !daemon {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT)

		_ = <-sig
		fmt.Println() //dead line to line up output

		// signal the processes
		notify.LnF("signaled sigusr1")
		procmcmd.Process.Signal(syscall.SIGUSR1)

		// shutdown the process manager
		time.Sleep(time.Second * 3)
		procmcmd.Process.Signal(syscall.SIGUSR2)
		procmcmd.Wait()
		notify.LnF("procm shutdown")

		// close the logs
		if procmlog != nil {
			procmlog.Close()
		}
		if corelog != nil {
			corelog.Close()
		}
		notify.LnF("shutdown complete")
	}
}

func startProcm(nolog bool) error {

	procmcmd = exec.Command("gmbhProcm")
	procmcmd.Env = append(os.Environ(), "ENV=M")

	if !nolog {
		var err error
		procmlog, err = getLogFile(config.LogPath, config.ProcmLogName)
		if err == nil {
			procmcmd.Stdout = procmlog
			procmcmd.Stderr = procmlog
			notify.LnF(filepath.Join(notify.Getpwd(), config.LogPath, config.ProcmLogName))
		} else {
			notify.LnRedF("could not create procm log, err=%s", err.Error())
		}
	}

	return procmcmd.Start()
}

func startCore(nolog, verbose bool) error {

	corecmd = exec.Command("gmbhProcm", "--remote", "--config=./gmbh/"+coreService)
	corecmd.Env = append(os.Environ(), "ENV=M")

	if !nolog {
		if verbose {
			corecmd.Args = append(corecmd.Args, "--verbose")
			corecmd.Stdout = os.Stdout
			corecmd.Stderr = os.Stderr
		} else {
			var err error
			corelog, err = getLogFile(config.LogPath, config.ProcmLogName)
			if err == nil {
				corecmd.Stdout = corelog
				corecmd.Stderr = corelog
				notify.LnF(filepath.Join(notify.Getpwd(), config.LogPath, config.ProcmLogName))
			} else {
				notify.LnRedF("could not create core log, err=%s", err.Error())
			}
		}
	}

	return corecmd.Start()
}

func startServices(conf *config.SystemConfig) {

	numNodes := math.Ceil(float64(len(conf.Service)) / float64(conf.MaxPerNode))

	for i := 0; i < int(numNodes); i++ {
		start := i * conf.MaxPerNode
		end := start + conf.MaxPerNode
		if end > len(conf.Service) {
			end = len(conf.Service)
		}

		genNode(i+1, conf.Service[start:end])

		err := launchService(i + 1)
		if err != nil {
			notify.LnRedF("error starting node %d; error=%s", i+1, err.Error())
		}
	}

}

func genNode(node int, services []*config.ServiceConfig) error {
	f, err := fileutil.CreateFile(filepath.Join("gmbh", "node_"+strconv.Itoa(node)+".toml"))
	if err != nil {
		return err
	}
	w := f.WriteString

	w(doNotEdit)
	w("#\n# Services in this file (by directory) \n")
	for _, s := range services {
		w(fmt.Sprintf("# - %s\n", fileutil.GetAbsFpath(s.BinPath)))
	}
	w("#\n# Fingerprint - the id that refers to this cluster\n")
	w(fmt.Sprintf("fingerprint = \"%s\"", fingerprint))

	for _, s := range services {

		w(fmt.Sprintf(
			service,
			s.ID,
			strArrtoStr(s.Args),
			s.Env,
			s.Language,
			s.BinPath,
			s.SrcPath,
		))
	}

	f.Close()
	return nil
}

func launchService(node int) error {

	cmd := exec.Command("gmbhProcm", "--remote", "--config=./gmbh/node_"+strconv.Itoa(node)+".toml")
	cmd.Env = append(os.Environ(), []string{
		"ENV=M",
		"FINGERPRINT=" + fingerprint,
		"PROJPATH=" + notify.Getpwd(),
	}...)

	f, err := getLogFile(config.LogPath, "node-"+strconv.Itoa(node)+".log")
	if err == nil {
		notify.LnF("%s", filepath.Join(notify.Getpwd(), config.LogPath, "node-"+strconv.Itoa(node)+".log"))
		cmd.Stdout = f
		cmd.Stderr = f
	} else {
		notify.LnRedF("could not create log file: " + err.Error())
	}

	return cmd.Start()
}

func strArrtoStr(arr []string) string {
	ret := "["
	for i, v := range arr {
		if v == "" {
			continue
		}
		ret += "\"" + v + "\""
		if i != len(arr)-1 {
			ret += ","
		}
	}
	ret += "]"
	return ret
}

// getLogFile attempts to add the desired path as an extension to the current
// directory as reported by os.GetWd(). The file is then opened or created
// and returned
func getLogFile(desiredPathExt, filename string) (*os.File, error) {
	// get pwd
	dir, err := os.Getwd()
	if err != nil {
		notify.LnRedF("getlogfile, pwd err=%s", err.Error())
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
		notify.LnRedF("openfile err=%s", err.Error())
		return nil, err
	}
	return file, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkInstall() bool {
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(config.ProcmBinPathMac); os.IsNotExist(err) {
			return false
		}
		if _, err := os.Stat(config.CoreBinPathMac); os.IsNotExist(err) {
			return false
		}
		return true
	} else if runtime.GOOS == "linux" {
		if _, err := os.Stat(config.ProcmBinPathLinux); os.IsNotExist(err) {
			return false
		}
		if _, err := os.Stat(config.CoreBinPathLinux); os.IsNotExist(err) {
			return false
		}
		return true
	}
	notify.LnRedF(fmt.Sprintf("OS support not implemented for %s", runtime.GOOS))
	return false
}

// basePath attempts to get the absolute path to the directory in which the config file is specified
func basePath(configPath string) string {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		notify.LnRedF("error=%v", err.Error())
		return ""
	}
	return filepath.Dir(abs)
}
