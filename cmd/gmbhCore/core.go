package main

/**
 * gcore.go
 * Abe Dick
 * January 2019
 */

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/grouter"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/router"
	"github.com/gmbh-micro/setting"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// The global config and controller for the core
// Not much of a way around this when using rpc
var core *Core

// CabalMode is the parameter that controls how data is sent between processes
type CabalMode int

const (
	// Ephemeral mode assumes that no enforcement of data types will be made at either
	// end of the gRPC calls between the gimlet-gmbh package while services exchange
	// data.
	//
	// Ephemeral mode works much like http handlers work in the http package of go.
	Ephemeral CabalMode = 0

	// Custom mode assumes that a shared structure will be used between both services.
	//
	// TODO: Semantics of how to enforce data modes
	Custom CabalMode = 1
)

// Core - internal representation of the gmbhCore core
type Core struct {
	Version           string
	CodeName          string
	ProjectPath       string
	Config            *setting.UserConfig
	Router            *router.Router
	log               *os.File
	logm              *sync.Mutex
	controlServerLock *sync.Mutex

	// old
	// ProjectConf    *ProjectConfig
	// serviceHandler *service.ServiceHandler
	router *grouter.Router
}

// tmpService is used during service discovery to hold tmp dat
type tmpService struct {
	path       string
	configFile string
}

// StartCore initializes settings of the core and creates the service handler and router
// needed to process requestes
func StartCore(path string, verbose bool, daemon bool) *Core {

	userConfig, err := setting.ParseUserConfig(path + defaults.CONFIG_FILE)
	if err != nil {
		panic(err)
	}

	core = &Core{
		Version:     defaults.VERSION,
		CodeName:    defaults.CODE,
		Config:      userConfig,
		Router:      router.NewRouter(),
		ProjectPath: path,

		// old
		// ServiceDir: "services",
		// verbose:      verbose,
		// daemon:       daemon,
		// CabalAddress: "localhost:59999",
		// CtrlAddress:  "localhost:59997",
		// serviceHandler: &service.ServiceHandler{
		// 	Services: make(map[string]*service.ServiceControl),
		// 	Names:    make([]string, 0),
		// },
		router: &grouter.Router{
			BaseAddress:    "localhost:",
			NextPortNumber: 40010,
			Localhost:      true,
		},
	}

	fmt.Println(*userConfig)

	notify.SetVerbose(verbose)
	// core.ProjectConf = core.parseProjectYamlConfig(path + defaults.CONFIG_FILE)
	core.logRuntimeData(path + defaults.SERVICE_LOG_PATH)
	return core
}

// ServiceDiscovery scans all directories in the ./services folder looking for gmbhCore configuration files
func (c *Core) ServiceDiscovery() {
	path := c.getServicePath()

	if !c.Config.Daemon {
		notify.StdMsgBlue("service discovery started in")
		notify.StdMsgBlue(path)
	} else {
		notify.StdMsgBlue("(3/3) service discovery in " + path)
	}

	notify.StdMsgDebug("service path: " + path)

	// scan the services directory and find all services
	servicePaths, err := c.scanForServices(path)
	if err != nil {
		notify.StdMsgErr(err.Error(), 1)
	}

	fmt.Println(servicePaths)

	for i, servicePath := range servicePaths {

		// Add service to router
		newService, err := c.Router.AddService(servicePath + defaults.CONFIG_FILE)
		if err != nil {
			notify.StdMsgErr(err.Error(), 1)
		}

		if !c.Config.Daemon {
			notify.StdMsgBlue(fmt.Sprintf("(%d/%d)", i+1, len(servicePaths)))
			notify.StdMsgBlue(newService.Static.Name, 1)
			notify.StdMsgBlue(servicePath, 1)
			if newService.Static.IsServer {
				notify.StdMsgBlue("assigning address: "+newService.Address, 1)
			}
		}

		// Start service
		pid, err := newService.StartService()
		if err != nil {
			notify.StdMsgErr(err.Error(), 1)
		}
		notify.StdMsgDebug(pid)

	}

	// tmpService := c.findAllServices(path)
	// for i, nservice := range tmpService {

	// staticData := c.parseYamlConfig(nservice.path + nservice.configFile)

	// if !c.Config.Daemon {
	// 	notify.StdMsgBlue(fmt.Sprintf("(%d/%d)", i+1, len(tmpService)))
	// 	notify.StdMsgBlue(staticData.Name, 1)
	// 	notify.StdMsgBlue(path+"/"+staticData.Path, 1)
	// }

	// newService := service.NewServiceControl(staticData)
	// err := c.serviceHandler.AddService(newService)
	// if err != nil {
	// 	notify.StdMsgErr("Could not add service")
	// 	notify.StdMsgErr("reported error: " + err.Error())
	// 	continue
	// }

	// if staticData.IsServer {
	// 	newService.Address = c.router.GetNextAddress()
	// 	if !c.Config.Daemon {
	// 		notify.StdMsgBlue("Assigning address: "+newService.Address, 1)
	// 	}
	// }

	// newService.BinPath = nservice.path + newService.Static.Path
	// newService.ConfigPath = nservice.path

	// pid, err := c.startService(newService)
	// if err != nil {
	// 	notify.StdMsgErr("Could not add service", 1)
	// 	notify.StdMsgErr("reported error: "+err.Error(), 1)
	// 	continue
	// }
	// if !c.Config.Daemon {
	// 	notify.StdMsgGreen(fmt.Sprintf("Service running in ephemeral mode with pid=(%v)", pid), 1)
	// } else {
	// 	notify.StdMsgBlue(fmt.Sprintf("(%d/%d) %v started in ephemeral mode with pid=(%v)", i+1, len(tmpService), newService.Name, pid), 0)
	// }
	// }

	go c.takeInventory()

	if !c.Config.Daemon {
		notify.StdMsgBlue("Startup complete")
		notify.StdMsgGreen("Blocking main thread until SIGINT")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	_ = <-sig

	if !c.Config.Daemon {
		notify.StdMsgMagenta("Recieved shutdown signal")
	}
	c.shutdown(false)
}

func (c *Core) getServicePath() string {
	return c.ProjectPath + "/" + c.Config.ServicesDirectory
}

// func (c *Core) startService(service *service.ServiceControl) (string, error) {

// 	if service.Static.Language == "go" {
// 		// fmt.Println("service config path: " + service.ConfigPath)
// 		service.Process = pmgmt.NewGoProcess(service.Name, service.BinPath, service.ConfigPath)
// 		pid, err := service.Process.Controller.Start(service.Process)
// 		if err != nil {
// 			return "", err
// 		}
// 		return strconv.Itoa(pid), nil
// 	}

// 	return "", nil
// }

// findAllServices looks for .yaml files in subdirectories of baseDir
// baseDir/dir/*.yaml
// TODO: Need to verify that we are getting the correct yaml file
// if there are several yaml files and if there are no yaml
// ALL ERRORS NEED TO BE HANDLED BETTER THAN WITH LOG.FATAL()
func (c *Core) findAllServices(baseDir string) []tmpService {
	services := []tmpService{}

	baseDirFiles, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Fatal(err)
	}

	// For every file in the baseDirectory
	for _, file := range baseDirFiles {

		// eval symbolicLinks first
		fpath := baseDir + "/" + file.Name()
		potentialSymbolic, err := filepath.EvalSymlinks(fpath)
		if err != nil {
			notify.StdMsgErr(err.Error(), 0)
			continue
		}

		// If it wasn't a symbolic path check if it was a dir, skip if not
		if fpath == potentialSymbolic {
			if !file.IsDir() {
				continue
			}
		}

		// Try and open the symbolic link path and check for dir, skip if not
		newFile, err := os.Stat(potentialSymbolic)
		if err != nil {
			notify.StdMsgErr(err.Error())
			continue
		}

		if !newFile.IsDir() {
			continue
		}

		// Looking through potential gmbH service directory
		serviceFiles, err := ioutil.ReadDir(baseDir + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		for _, sfile := range serviceFiles {
			match, err := regexp.MatchString(defaults.CONFIG_FILE_EXT, sfile.Name())
			if err == nil && match {
				newService := tmpService{
					path:       baseDir + "/" + file.Name() + "/",
					configFile: sfile.Name(),
				}
				services = append(services, newService)
			}
		}
	}

	return services
}

// scanForServices scans for directories (or symbolic links to directories)
// that containa gmbh config file and returns an array of absolute paths
// to any found directories that contain the config file
func (c *Core) scanForServices(baseDir string) ([]string, error) {
	servicePaths := []string{}

	baseDirFiles, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return servicePaths, errors.New("could not scan base directory: " + err.Error())
	}

	for _, file := range baseDirFiles {

		// eval symbolic links
		fpath := baseDir + "/" + file.Name()
		potentialSymbolic, err := filepath.EvalSymlinks(fpath)
		if err != nil {
			notify.StdMsgErr(err.Error(), 0)
			continue
		}

		// If it wasn't a symbolic path check if it was a dir, skip if not
		if fpath == potentialSymbolic {
			if !file.IsDir() {
				continue
			}
		}

		// Try and open the symbolic link path and check for dir, skip if not
		newFile, err := os.Stat(potentialSymbolic)
		if err != nil {
			notify.StdMsgErr(err.Error())
			continue
		}

		if !newFile.IsDir() {
			continue
		}

		// Looking through potential gmbH service directory
		serviceFiles, err := ioutil.ReadDir(baseDir + "/" + file.Name())
		if err != nil {
			log.Fatal(err)
		}

		for _, sfile := range serviceFiles {
			match, err := regexp.MatchString(defaults.CONFIG_FILE_EXT, sfile.Name())
			if err == nil && match {
				servicePaths = append(servicePaths, baseDir+file.Name())
			}
		}
	}

	return servicePaths, nil
}

// func (c *Core) parseYamlConfig(path string) *service.StaticControl {
// 	var static service.StaticControl
// 	yamlFile, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		log.Printf("yamlFile.Get err   #%v ", err)
// 	}
// 	err = yaml.Unmarshal(yamlFile, &static)
// 	if err != nil {
// 		log.Fatalf("Unmarshal: %v", err)
// 	}
// 	return &static
// }

// StartCabalServer starts the gRPC server to run core on
func (c *Core) StartCabalServer() {
	go func() {
		list, err := net.Listen("tcp", c.Config.DefaultHost+c.Config.DefaultPort)
		if err != nil {
			panic(err)
		}

		s := grpc.NewServer()
		cabal.RegisterCabalServer(s, &cabalServer{})

		reflection.Register(s)
		if err := s.Serve(list); err != nil {
			panic(err)
		}

	}()
	if !c.Config.Daemon {
		notify.StdMsgBlue("attempting to start cabal server")
		notify.StdMsgGreen("starting cabal server at "+c.Config.DefaultHost+c.Config.DefaultPort, 1)
	} else {
		notify.StdMsgBlue("(1/3) starting cabal server at " + c.Config.DefaultHost + c.Config.DefaultPort)
	}
}

// StartControlServer starts the gRPC server to run core on
func (c *Core) StartControlServer() {
	go func() {
		list, err := net.Listen("tcp", c.Config.ControlHost+c.Config.ControlPort)
		if err != nil {
			panic(err)
		}

		s := grpc.NewServer()
		cabal.RegisterControlServer(s, &controlServer{})

		reflection.Register(s)
		if err := s.Serve(list); err != nil {
			panic(err)
		}

	}()
	if !c.Config.Daemon {
		notify.StdMsgBlue("Attempting to start control server")
		notify.StdMsgGreen("starting control server at "+c.Config.ControlHost+c.Config.ControlPort, 1)
	} else {
		notify.StdMsgBlue("(2/3) starting control server at " + c.Config.ControlHost + c.Config.ControlPort)
	}
}

func (c *Core) logRuntimeData(path string) {
	filename := ".gmbhCore"
	var err error
	c.log, err = notify.OpenLogFile(path, filename)
	c.logm = &sync.Mutex{}
	if err != nil {
		notify.StdMsgErr("could not create log file: " + err.Error())
		return
	}
	sep := "------------------------------------------------------------------------"
	c.log.WriteString("\n" + sep + "\n")
	c.log.WriteString("startTime=\"" + time.Now().Format("Jan 2 2006 15:04:05 MST") + "\"\n")
	c.log.WriteString("cabalAddress=\"" + c.Config.DefaultHost + c.Config.DefaultPort + "\"\n")
	c.log.WriteString("ctrlAddress=\"" + c.Config.ControlHost + c.Config.ControlPort + "\"\n")

}

func (c *Core) takeInventory() {
	// serviceString := "services=["
	// for _, serviceName := range c.serviceHandler.Names {
	// 	service := c.serviceHandler.Services[serviceName]
	// 	serviceString += "\"" + service.Name + "-" + service.ConfigPath + "\", "
	// }
	// serviceString = serviceString[:len(serviceString)-2]
	// serviceString += "]"
	// c.logm.Lock()
	// defer c.logm.Unlock()
	// c.log.WriteString(serviceString + "\n")
}

func (c *Core) shutdown(remote bool) {
	defer os.Exit(0)
	if remote {
		if !c.Config.Daemon {
			notify.StdMsgGreen("Recieved remote shutdown notification")
		}
		time.Sleep(time.Second * 2)
	}
	c.Router.KillAllServices()
	c.logm.Lock()
	c.log.WriteString("stopTime=\"" + time.Now().Format("Jan 2 2006 15:04:05 MST") + "\"\n")
	c.logm.Unlock()
	c.log.Close()
}

func getCore() (*Core, error) {
	if core == nil {
		return nil, errors.New("could not find core instance")
	}
	return core, nil
}
