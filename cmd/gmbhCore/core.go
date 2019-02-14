package main

// /**
//  * gcore.go
//  * Abe Dick
//  * January 2019
//  */

// import (
// 	"errors"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"path/filepath"
// 	"regexp"
// 	"strings"
// 	"syscall"
// 	"time"

// 	"github.com/gmbh-micro/cabal"
// 	"github.com/gmbh-micro/defaults"
// 	"github.com/gmbh-micro/notify"
// 	"github.com/gmbh-micro/router"
// 	"github.com/gmbh-micro/rpc"
// 	"github.com/gmbh-micro/service"
// 	"github.com/gmbh-micro/service/process"
// 	"github.com/gmbh-micro/service/static"
// 	yaml "gopkg.in/yaml.v2"
// )

// // mode changes how things are allowed to be attached to Core
// // whether it be open to new services or closed to a manifest file
// type mode int

// const (
// 	local mode = 1
// )

// // The global config and controller for the core
// // Not much of a way around this when using rpc
// var core *Core

// // Core - internal representation of the gmbhCore core
// type Core struct {
// 	Version     string
// 	Mode        mode
// 	CodeName    string
// 	ProjectPath string
// 	MsgCounter  int
// 	StartTime   time.Time
// 	Log         *notify.Log
// 	acclog      *notify.Log
// 	errlog      *notify.Log
// 	Config      *UserConfig
// 	Cabal       *rpc.Connection
// 	Control     *rpc.Connection
// 	Router      *router.Router
// }

// // StartCore initializes settings of the core and creates the service handler and router
// // needed to process requestes
// func StartCore(path string) (*Core, error) {

// 	userConfig, err := ParseUserConfig(path + defaults.CONFIG_FILE)
// 	if err != nil {
// 		return nil, err
// 	}
// 	userConfig.Daemon = false
// 	core = &Core{
// 		Version:     defaults.VERSION,
// 		CodeName:    defaults.CODE,
// 		Config:      userConfig,
// 		Router:      router.NewRouter(),
// 		ProjectPath: path,
// 		MsgCounter:  1,
// 		Log:         notify.NewLogFile(path+defaults.LOG_PATH, defaults.CORE_LOG_NAME, false),
// 		acclog:      notify.NewLogFile(path+defaults.LOG_PATH, defaults.ACCESS_LOG_NAME, false),
// 		errlog:      notify.NewLogFile(path+defaults.LOG_PATH, defaults.ERROR_LOG_NAME, false),
// 		StartTime:   time.Now(),
// 	}
// 	core.Log.Sep()
// 	core.Log.Ln("                    _           ")
// 	core.Log.Ln("  _  ._ _  |_  |_| /   _  ._ _  ")
// 	core.Log.Ln(" (_| | | | |_) | | \\_ (_) | (/_")
// 	core.Log.Ln("  _|                            ")
// 	core.Log.Ln("version=%v; code=%v; startTime=%s", core.Version, core.CodeName, core.StartTime.Format(time.Stamp))

// 	return core, nil
// }

// func getCore() (*Core, error) {
// 	if core == nil {
// 		return nil, errors.New("could not find core instance")
// 	}
// 	return core, nil
// }

// // ServiceDiscovery scans all directories in the ./services folder looking for gmbhCore configuration files
// func (c *Core) ServiceDiscovery() error {

// 	path := c.ProjectPath + "/" + c.Config.ServicesDirectory
// 	// c.Log.Ln("Managed Services (path = %s)", path)

// 	// scan the services directory and find all services
// 	servicePaths, err := c.scanForServices(path)
// 	if err != nil {
// 		return errors.New("could not read service directory")
// 	}

// 	// Create and attach all services that run in Managed mode
// 	for i, servicePath := range servicePaths {

// 		// Add a new managed service to router
// 		newService, err := c.Router.AddManagedService(servicePath + defaults.CONFIG_FILE)
// 		if err != nil {
// 			c.Log.Err("(%d/%d) error = %s", i+1, len(servicePaths), err.Error())
// 			continue
// 		}

// 		// Start service
// 		pid, err := newService.Start()
// 		if err != nil {
// 			c.Log.Err("(%d/%d) error = %s", i+1, len(servicePaths), err.Error())
// 			continue
// 		}

// 		c.Log.Ln("(%d/%d) name = %s", i+1, len(servicePaths), newService.Static.Name)
// 		c.Log.Ln("      path = %s\n      address = %s\n      pid = %s", servicePath, newService.Address, pid)
// 	}

// 	return nil
// }

// // Wait for shutdown
// func (c *Core) Wait() {
// 	sig := make(chan os.Signal, 1)
// 	signal.Notify(sig, syscall.SIGINT)
// 	_ = <-sig

// 	c.shutdown(false)
// }

// // scanForServices scans for directories (or symbolic links to directories)
// // that containa gmbh config file and returns an array of absolute paths
// // to any found directories that contain the config file
// // TODO: Need to verify that we are getting the correct yaml file
// // if there are several yaml files and if there are no yaml
// func (c *Core) scanForServices(baseDir string) ([]string, error) {
// 	servicePaths := []string{}

// 	baseDirFiles, err := ioutil.ReadDir(baseDir)
// 	if err != nil {
// 		return servicePaths, errors.New("could not scan base directory: " + err.Error())
// 	}

// 	for _, file := range baseDirFiles {

// 		// eval symbolic links
// 		fpath := baseDir + "/" + file.Name()
// 		potentialSymbolic, err := filepath.EvalSymlinks(fpath)
// 		if err != nil {
// 			notify.StdMsgErr(err.Error(), 0)
// 			continue
// 		}

// 		// If it wasn't a symbolic path check if it was a dir, skip if not
// 		if fpath == potentialSymbolic {
// 			if !file.IsDir() {
// 				continue
// 			}
// 		}

// 		// Try and open the symbolic link path and check for dir, skip if not
// 		newFile, err := os.Stat(potentialSymbolic)
// 		if err != nil {
// 			notify.StdMsgErr(err.Error())
// 			continue
// 		}

// 		if !newFile.IsDir() {
// 			continue
// 		}

// 		// Looking through potential gmbH service directory
// 		serviceFiles, err := ioutil.ReadDir(baseDir + "/" + file.Name())
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		for _, sfile := range serviceFiles {
// 			match, err := regexp.MatchString(defaults.CONFIG_FILE_EXT, sfile.Name())
// 			if err == nil && match {
// 				servicePaths = append(servicePaths, baseDir+file.Name())
// 			}
// 		}
// 	}

// 	return servicePaths, nil
// }

// // StartCabalServer starts the gRPC server to run core on
// func (c *Core) StartCabalServer() error {
// 	c.Cabal = rpc.NewCabalConnection(c.Config.DefaultHost+c.Config.DefaultPort, &cabalServer{})
// 	err := c.Cabal.Connect()
// 	if err != nil {
// 		c.Log.Err("could not start cabal server\nerror = %v", err.Error())
// 		return err
// 	}
// 	c.Log.Ln("cabal server started at %v", c.Config.DefaultHost+c.Config.DefaultPort)
// 	return nil
// }

// // StartControlServer starts the gRPC server to run core on
// func (c *Core) StartControlServer() error {
// 	c.Control = rpc.NewControlConnection(c.Config.ControlHost+c.Config.ControlPort, &controlServer{})
// 	err := c.Control.Connect()
// 	if err != nil {
// 		c.Log.Err("could not start control server\nerror = %v", err.Error())
// 		return err
// 	}
// 	c.Log.Ln("control server started at %v", c.Config.ControlHost+c.Config.ControlPort)
// 	return nil
// }

// func (c *Core) shutdown(remote bool) {
// 	if remote {
// 		time.Sleep(time.Second * 2)
// 	}

// 	// Send shutdown notice to all managed services
// 	c.Router.KillManagedServices()

// 	// Send shutdown notice to all remote services
// 	for _, rs := range c.Router.GetAllRemoteServices() {
// 		fmt.Println("sending shutdown to: ", rs.Static.Name)
// 		c.sendServiceShutdownNotice(rs)
// 	}

// 	// send shutdown notice to all process managers
// 	for _, rc := range c.Router.GetAllProcessManagers() {
// 		c.sendProcessManagerShutdownNotice(rc)
// 	}

// 	c.Cabal.Disconnect()
// 	c.Control.Disconnect()

// 	c.Log.Ln("shutdownTime = %v; duration=%v", time.Now().Format(time.Stamp), time.Since(c.StartTime))
// 	os.Exit(0)
// }

// func (c *Core) sendServiceShutdownNotice(serv *service.Service) {
// 	if serv.Address == "" {
// 		notify.StdMsgErr("core.sendServiceShutdownNotice.cannotFind=" + serv.Static.Name)
// 		return
// 	}
// 	client, ctx, can, err := rpc.GetCabalRequest(serv.Address, time.Second)
// 	if err != nil {
// 		notify.StdMsgErr("core.sendServiceShutdownNotice.cannotContact=" + serv.Static.Name)
// 		return
// 	}
// 	defer can()
// 	request := &cabal.ServiceUpdate{
// 		Sender:  "core",
// 		Target:  serv.Static.Name,
// 		Message: "core shutdown",
// 		Action:  "core.shutdown",
// 	}
// 	client.UpdateServiceRegistration(ctx, request)
// }

// func (c *Core) sendProcessManagerShutdownNotice(cont *process.RemoteManager) {
// 	if cont.Address == "" {
// 		notify.StdMsgErr("core.sendProcessManageprShutdownNotice.cannotFind=" + cont.ID)
// 	}
// 	client, ctx, can, err := rpc.GetRemoteRequest(cont.Address, time.Second)
// 	if err != nil {
// 		notify.StdMsgErr("core.sendServiceShutdownNotice.cannotContact=" + cont.ID)
// 		return
// 	}
// 	defer can()
// 	request := &cabal.ServiceUpdate{
// 		Sender:  "core",
// 		Target:  cont.ID,
// 		Message: "core shutdown",
// 		Action:  "core.shutdown",
// 	}
// 	client.UpdateServiceRegistration(ctx, request)
// }

// func (c *Core) registerPlanetaryService(name string, aliases []string, isclient bool, isserver bool) (*service.Service, error) {
// 	static := &static.Static{
// 		Name:     name,
// 		Aliases:  aliases,
// 		Mode:     "planetary",
// 		IsClient: isclient,
// 		IsServer: isserver,
// 	}
// 	return c.Router.AddPlanetaryService(static)
// }

// // GetMsgCount of the current msg counter and increment the count
// func (c *Core) GetMsgCount() int {
// 	defer func() { c.MsgCounter++ }()
// 	return c.MsgCounter
// }

// // RequestHandler ; handle making requests between services in the raw protobuffer objects
// type RequestHandler struct {
// 	Request   *cabal.Request
// 	Responder *cabal.Responder
// 	Errors    []error
// 	startTime time.Time
// 	success   bool
// 	count     int
// }

// func newRequestHandler(request *cabal.Request) RequestHandler {
// 	return RequestHandler{
// 		Request:   request,
// 		startTime: time.Now(),
// 	}
// }

// // Fulfill 's the request
// func (r *RequestHandler) Fulfill() {

// 	// Get the core instance
// 	c, err := getCore()
// 	if err != nil {
// 		r.Errors = append(r.Errors, err)
// 		r.reportRequest(c.Config.Daemon)
// 		return
// 	}

// 	r.count = c.GetMsgCount()

// 	// Get the address of the target from the router
// 	address, err := c.Router.LookupServiceAddress(r.Request.Target)
// 	if err != nil {
// 		r.Errors = append(r.Errors, err)
// 		r.processErrors()
// 		r.reportRequest(c.Config.Daemon)
// 		return
// 	}

// 	// make sure that the address exists
// 	if address == "" {
// 		r.Errors = append(r.Errors, errors.New("requestHandler.Fulfill.addressStringEmpty"))
// 		r.processErrors()
// 		r.reportRequest(c.Config.Daemon)
// 		return
// 	}

// 	r.success = true

// 	// forward the message and get the responder
// 	r.forewardRequest(address)

// 	// process the errors to make sure that there is always a responder and it always
// 	// is reporting the correct error data
// 	r.processErrors()

// 	// Notify stdOut or logger of request
// 	r.reportRequest(c.Config.Daemon)
// }

// func (r *RequestHandler) processErrors() {

// 	if r.Responder == nil {
// 		r.Responder = &cabal.Responder{}
// 	}

// 	if len(r.Errors) > 0 {
// 		r.Responder.HadError = true
// 		errStrs := []string{}
// 		for _, e := range r.Errors {
// 			errStrs = append(errStrs, e.Error())
// 		}
// 		r.Responder.ErrorString = strings.Join(errStrs, ",")
// 	}
// }

// func (r *RequestHandler) forewardRequest(address string) {
// 	client, ctx, can, err := rpc.GetCabalRequest(address, time.Second)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer can()

// 	request := cabal.DataReq{Req: r.Request}

// 	response, err := client.MakeDataRequest(ctx, &request)
// 	if err != nil {
// 		r.Errors = append(r.Errors, err)
// 		return
// 	}
// 	r.Responder = response.GetResp()
// }

// // GetResponder to fulfil transaction between services
// func (r *RequestHandler) GetResponder() *cabal.Responder {
// 	return r.Responder
// }

// func (r *RequestHandler) reportRequest(silently bool) {
// 	if !silently {
// 		notify.StdMsgBlueNoPrompt(fmt.Sprintf("[data] <(%d)- sender=(%s); target=(%s); method=(%s)", r.count, r.Request.Sender, r.Request.Target, r.Request.Method))
// 		if r.success == true {
// 			notify.StdMsgBlueNoPrompt(fmt.Sprintf("       -(%d)> Success; duration=(%s); errorString=(%s)", r.count, time.Since(r.startTime), r.Responder.ErrorString))
// 			return
// 		}
// 		notify.StdMsgErrNoPrompt(fmt.Sprintf("       -(%d)> Failed; duration=(%s); errorString=(%s)", r.count, time.Since(r.startTime), r.Responder.ErrorString))
// 		return
// 	}
// }

// // UserConfig represents the parsable config settings
// type UserConfig struct {
// 	Name              string   `yaml:"project_name"`
// 	Verbose           bool     `yaml:"verbose"`
// 	Daemon            bool     `yaml:"daemon"`
// 	DefaultHost       string   `yaml:"default_host"`
// 	DefaultPort       string   `yaml:"default_port"`
// 	ControlHost       string   `yaml:"control_host"`
// 	ControlPort       string   `yaml:"control_port"`
// 	ServicesDirectory string   `yaml:"services_directory"`
// 	ServicesToAttach  []string `yaml:"services_to_attach"`
// 	ServicesDetached  []string `yaml:"services_detached"`
// }

// // ParseUserConfig attempts to parse a yaml file at path and return the UserConfigStruct.
// // If not all settings have been defined in user path, the defaults will be used.
// func ParseUserConfig(path string) (*UserConfig, error) {
// 	c := UserConfig{Verbose: defaults.VERBOSE, Daemon: defaults.DAEMON}

// 	yamlFile, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return nil, errors.New("could not open yaml file: " + err.Error())
// 	}

// 	err = yaml.Unmarshal(yamlFile, &c)
// 	if err != nil {
// 		return nil, errors.New("could not parse yaml file: " + err.Error())
// 	}

// 	if c.Name == "" {
// 		c.Name = defaults.PROJECT_NAME
// 	}
// 	if c.DefaultHost == "" {
// 		c.DefaultHost = defaults.DEFAULT_HOST
// 	}
// 	if c.DefaultPort == "" {
// 		c.DefaultPort = defaults.DEFAULT_PORT
// 	}
// 	if c.ControlHost == "" {
// 		c.ControlHost = defaults.CONTROL_HOST
// 	}
// 	if c.ControlPort == "" {
// 		c.ControlPort = defaults.CONTROL_PORT
// 	}
// 	return &c, nil
// }
