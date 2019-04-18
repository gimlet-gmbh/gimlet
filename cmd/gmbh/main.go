package main

import (
	"encoding/json"
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

const coreService = "core.toml"

var fingerprint = xid.New().String()

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
			print("error, must specify a gmbh config file")
			os.Exit(1)
		}

		if *run {
			start(*config, *verbose, *nolog, *daemon)
		} else {
			builddeploy(*config, *verbose)
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

	print("                  ")
	print("  _  ._ _  |_  |_|")
	print(" (_| | | | |_) | |")
	print("  _|              ")
	print("Version=%s; Code=%s", config.Version, config.Code)

	// make sure that gmbhCore is installed
	installed := checkInstall()
	if !installed {
		print("gmbhCore or gmbhProcm does not seem to be installed")
		os.Exit(1)
	}

	conf, err := config.ParseSystemConfig(cfile)
	if err != nil {
		print("specified config file cannot be parsed, err=%s", err.Error())
		os.Exit(1)
	}

	fileutil.MkDir("gmbh")

	if !fileutil.FileExists(filepath.Join("gmbh", coreService)) {
		print("Generating core service config file...")
		err = genCoreConf(filepath.Join("gmbh", coreService), cfile, conf)
		if err != nil {
			print("cannot create core config file, err=%s", err.Error())
			os.Exit(1)
		}
	}

	proccmd, proclog, err := startProcm(nolog)
	if err != nil {
		print("could not start ProcM, err=%s", err.Error())
		os.Exit(1)
	}

	_, corelog, err := startCore(nolog, verbose)
	if err != nil {
		print("could not start core, err=%s", err.Error())
		os.Exit(1)
	}

	startServices(conf)

	if !daemon {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT)

		_ = <-sig
		fmt.Println() //dead line to line up output

		// signal the processes
		// print("signaled sigusr1")
		proccmd.Process.Signal(syscall.SIGALRM)

		// shutdown the process manager
		proccmd.Process.Signal(syscall.SIGQUIT)
		proccmd.Wait()
		// print("procm shutdown")

		// close the logs
		if proclog != nil {
			proclog.Close()
		}
		if corelog != nil {
			corelog.Close()
		}
		print("shutdown complete...")
	}
}

func startProcm(nolog bool) (*exec.Cmd, *os.File, error) {

	cmd := exec.Command("gmbhProcm")
	cmd.Env = append(os.Environ(), "ENV=M")

	var log *os.File
	var err error
	if !nolog {
		log, err = fileutil.GetLogFile(config.LogPath, config.ProcmLogName)
		if err == nil {
			cmd.Stdout = log
			cmd.Stderr = log
			print(filepath.Join(fileutil.Getpwd(), config.LogPath, config.ProcmLogName))
		} else {
			print("could not create procm log, err=%s", err.Error())
		}
	}

	return cmd, log, cmd.Start()
}

func startCore(nolog, verbose bool) (*exec.Cmd, *os.File, error) {

	cmd := exec.Command("gmbhProcm", "--remote", "--config=./gmbh/"+coreService)

	_, f, err := config.ParseServices("./gmbh/" + coreService)
	if err != nil {
		print("could not parse core service config; err=%s", err.Error())
		os.Exit(1)
	}
	cmd.Env = append(os.Environ(), []string{"ENV=M", "FINGERPRINT=" + f}...)

	var log *os.File
	if !nolog {
		if verbose {
			cmd.Args = append(cmd.Args, "--verbose")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			log, err = fileutil.GetLogFile(config.LogPath, config.ProcmLogName)
			if err == nil {
				cmd.Stdout = log
				cmd.Stderr = log
				print(filepath.Join(fileutil.Getpwd(), config.LogPath, config.ProcmLogName))
			} else {
				print("could not create core log, err=%s", err.Error())
			}
		}
	}

	return cmd, log, cmd.Start()
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
			print("error starting node %d; error=%s", i+1, err.Error())
		}
	}

}

func launchService(node int) error {

	cmd := exec.Command("gmbhProcm", "--remote", "--config=./gmbh/node_"+strconv.Itoa(node)+".toml")
	cmd.Env = append(os.Environ(), []string{
		"ENV=M",
		"FINGERPRINT=" + fingerprint,
		"PROJPATH=" + fileutil.Getpwd(),
	}...)

	f, err := fileutil.GetLogFile(config.LogPath, "node-"+strconv.Itoa(node)+".log")
	if err == nil {
		print("%s", filepath.Join(fileutil.Getpwd(), config.LogPath, "node-"+strconv.Itoa(node)+".log"))
		cmd.Stdout = f
		cmd.Stderr = f
	} else {
		print("could not create log file: " + err.Error())
	}

	return cmd.Start()
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
		args, _ := json.Marshal(s.Args)
		w(fmt.Sprintf(service, s.ID, args, s.Env, s.Language, s.BinPath, s.SrcPath))
	}

	f.Close()
	return nil
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
	} else if runtime.GOOS == "windows" {
		if _, err := os.Stat(config.ProcmBinPathWindows); os.IsNotExist(err) {
			return false
		}
		if _, err := os.Stat(config.CoreBinPathWindows); os.IsNotExist(err) {
			return false
		}
		notify.LnRedF("Windows support is incomplete")
		return true
	}
	print(fmt.Sprintf("OS support not implemented for %s", runtime.GOOS))
	return false
}

func print(format string, a ...interface{}) {
	format = "[" + time.Now().Format(config.LogStamp) + "] [gmbh] " + format
	notify.LnYellowF(format, a...)
}
