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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/router"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/static"
	"github.com/gmbh-micro/setting"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// The global config and controller for the core
// Not much of a way around this when using rpc
var core *Core

// Core - internal representation of the gmbhCore core
type Core struct {
	Version           string
	CodeName          string
	ProjectPath       string
	Config            *setting.UserConfig
	Router            *router.Router
	MsgCounter        int
	log               *os.File
	logm              *sync.Mutex
	controlServerLock *sync.Mutex
}

// StartCore initializes settings of the core and creates the service handler and router
// needed to process requestes
func StartCore(path string, verbose bool, daemon bool) *Core {

	userConfig, err := setting.ParseUserConfig(path + defaults.CONFIG_FILE)
	if err != nil {
		panic(err)
	}
	userConfig.Daemon = daemon
	core = &Core{
		Version:     defaults.VERSION,
		CodeName:    defaults.CODE,
		Config:      userConfig,
		Router:      router.NewRouter(),
		ProjectPath: path,
		MsgCounter:  1,
	}

	notify.SetVerbose(verbose)
	core.logRuntimeData(path + defaults.SERVICE_LOG_PATH)
	return core
}

func getCore() (*Core, error) {
	if core == nil {
		return nil, errors.New("could not find core instance")
	}
	return core, nil
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

	// scan the services directory and find all services
	servicePaths, err := c.scanForServices(path)
	if err != nil {
		notify.StdMsgErr(err.Error(), 1)
	}

	// Create and attach all services that run in Managed mode
	for i, servicePath := range servicePaths {

		// Add a new managed service to router
		newService, err := c.Router.AddManagedService(servicePath + defaults.CONFIG_FILE)
		if err != nil {
			// report the error and skip the rest for now
			// TODO: Better process error handling
			notify.StdMsgBlue(fmt.Sprintf("(%d/%d)", i+1, len(servicePaths)))
			notify.StdMsgErr(err.Error(), 1)
			continue
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
		if !c.Config.Daemon {
			notify.StdMsgGreen(fmt.Sprintf("Service running in ephemeral mode with pid=(%v)", pid), 1)
		} else {
			notify.StdMsgBlue(fmt.Sprintf("(%d/%d) %v started in ephemeral mode with pid=(%v)", i+1, len(servicePaths), newService.Static.Name, pid), 0)
		}

	}

	notify.StdMsgBlue("startup complete")
	if !c.Config.Daemon {
		notify.StdMsgGreen("Blocking main thread until SIGINT")
	}

	go c.takeInventory()

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

// scanForServices scans for directories (or symbolic links to directories)
// that containa gmbh config file and returns an array of absolute paths
// to any found directories that contain the config file
// TODO: Need to verify that we are getting the correct yaml file
// if there are several yaml files and if there are no yaml
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
	paths := c.Router.TakeInventory()
	result := "services=[" + strings.Join(paths, ", ") + "]\n"
	c.logm.Lock()
	defer c.logm.Unlock()
	c.log.WriteString(result)
}

func (c *Core) shutdown(remote bool) {
	defer os.Exit(0)
	if remote {
		if !c.Config.Daemon {
			notify.StdMsgGreen("n Recieved remote shutdown notification")
		}
		time.Sleep(time.Second * 2)
	}
	c.Router.KillAllServices()
	c.logm.Lock()
	c.log.WriteString("stopTime=\"" + time.Now().Format("Jan 2 2006 15:04:05 MST") + "\"\n")
	c.logm.Unlock()
	c.log.Close()
}

func (c *Core) registerRemoteService(name string, aliases []string, isclient bool, isserver bool) (*service.Service, error) {
	static := &static.Static{
		Name:     name,
		Aliases:  aliases,
		Mode:     "remote",
		IsClient: isclient,
		IsServer: isserver,
	}
	return c.Router.AddRemoteService(static)
}

// GetMsgCount of the current msg counter and increment the count
func (c *Core) GetMsgCount() int {
	defer func() { c.MsgCounter++ }()
	return c.MsgCounter
}

// RequestHandler ; handle making requests between services in the raw protobuffer objects
type RequestHandler struct {
	Request   *cabal.Request
	Responder *cabal.Responder
	Errors    []error
	startTime time.Time
	success   bool
	count     int
}

func newRequestHandler(request *cabal.Request) RequestHandler {
	return RequestHandler{
		Request:   request,
		startTime: time.Now(),
	}
}

// Fulfill 's the request
func (r *RequestHandler) Fulfill() {

	// Get the core instance
	c, err := getCore()
	if err != nil {
		r.Errors = append(r.Errors, err)
		r.reportRequest(c.Config.Daemon)
		return
	}

	r.count = c.GetMsgCount()

	// Get the address of the target from the router
	address, err := c.Router.LookupAddress(r.Request.Target)
	if err != nil {
		r.Errors = append(r.Errors, err)
		r.processErrors()
		r.reportRequest(c.Config.Daemon)
		return
	}

	// make sure that the address exists
	if address == "" {
		r.Errors = append(r.Errors, errors.New("requestHandler.Fulfill.addressStringEmpty"))
		r.processErrors()
		r.reportRequest(c.Config.Daemon)
		return
	}

	r.success = true

	// forward the message and get the responder
	r.forewardRequest(address)

	// process the errors to make sure that there is always a responder and it always
	// is reporting the correct error data
	r.processErrors()

	// Notify stdOut or logger of request
	r.reportRequest(c.Config.Daemon)
}

func (r *RequestHandler) processErrors() {

	if r.Responder == nil {
		r.Responder = &cabal.Responder{}
	}

	if len(r.Errors) > 0 {
		r.Responder.HadError = true
		errStrs := []string{}
		for _, e := range r.Errors {
			errStrs = append(errStrs, e.Error())
		}
		r.Responder.ErrorString = strings.Join(errStrs, ",")
	}
}

func (r *RequestHandler) forewardRequest(address string) {
	client, ctx, can, err := makeRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	request := cabal.DataReq{Req: r.Request}

	response, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		r.Errors = append(r.Errors, err)
		return
	}
	r.Responder = response.GetResp()
}

// GetResponder to fulfil transaction between services
func (r *RequestHandler) GetResponder() *cabal.Responder {
	return r.Responder
}

func (r *RequestHandler) reportRequest(silently bool) {
	if !silently {
		notify.StdMsgBlue(fmt.Sprintf("<(%d)- Processing data request; sender=(%s); target=(%s); method=(%s)", r.count, r.Request.Sender, r.Request.Target, r.Request.Method))
		if r.success == true {
			notify.StdMsgGreen(fmt.Sprintf("-(%d)> Success; duration=(%s); errorString=(%s)", r.count, time.Since(r.startTime), r.Responder.ErrorString))
			return
		}
		notify.StdMsgErr(fmt.Sprintf("-(%d)> Failed; duration=(%s); errorString=(%s)", r.count, time.Since(r.startTime), r.Responder.ErrorString))
		return
	}
}
