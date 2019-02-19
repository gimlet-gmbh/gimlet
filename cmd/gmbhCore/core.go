package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"github.com/rs/xid"
	"google.golang.org/grpc/metadata"
)

// internal reference to core for use rpc
var core *Core

// Core is the main gmbh controller
type Core struct {
	Version string
	Code    string

	// the filesystem directory to the gmbh project where assumptions can be made about
	// structure accoring to the config file
	ProjectPath string

	// con holds the host connection for the cabal server
	con *rpc.Connection

	// conf is the user configurable parameters as read in from file
	conf *config.SystemCore

	// Router controls all aspects of data requests & handling in Core
	Router *Router

	// mode can be set to managed to indicate that the service launcher
	// has launched core and it is being controlled by a remote
	mode string

	// parent id is for remote instances of core
	parentID string

	msgCounter  int
	startTime   time.Time
	mu          *sync.Mutex
	verbose     bool
	verboseData bool
}

// NewCore initializes settings of the core and instantiates the core struct which includes the
// service router and handlers
func NewCore(cPath string, verbose, verboseData bool) (*Core, error) {

	// cannot reinit core once it has been created
	if core != nil {
		return core, nil
	}

	userConfig, err := config.ParseSystemCore(cPath)
	if err != nil {
		notify.LnRedF("could not parse config; err=%v", err.Error())
		return nil, err
	}

	core = &Core{
		Version:     config.Version,
		Code:        config.Code,
		ProjectPath: basePath(cPath),
		con:         rpc.NewCabalConnection(userConfig.Address, &cabalServer{}),
		conf:        userConfig,
		Router:      NewRouter(),
		msgCounter:  1,
		startTime:   time.Now(),
		mode:        os.Getenv("SERVICEMODE"),
		parentID:    os.Getenv("REMOTE"),
		mu:          &sync.Mutex{},
		verbose:     verbose,
	}

	if core.ProjectPath == "" {
		notify.LnRedF("could not calculate path to project")
		return nil, errors.New("config path error")
	}

	notify.LnCyanF("                    _            _              ")
	notify.LnCyanF("  _  ._ _  |_  |_| /   _  ._ _  | \\  _. _|_  _. ")
	notify.LnCyanF(" (_| | | | |_) | | \\_ (_) | (/_ |_/ (_|  |_ (_| ")
	notify.LnCyanF("  _|                                            ")
	notify.LnCyanF("version=%v; code=%v; startTime=%s", core.Version, core.Code, core.startTime.Format(time.Stamp))

	return core, nil
}

// GetCore retrieves the instance of core. For use with rpc server
func GetCore() (*Core, error) {
	if core != nil {
		return core, nil
	}
	return nil, errors.New("core.GetCore.internalError")
}

// Start the cabal server
func (c *Core) Start() {
	err := c.con.Connect()
	if err != nil {
		c.ve("could not connected; err=%s", err.Error())
		return
	}
	c.v("connected; address=%s", c.con.Address)

	// c.serviceDiscovery()

	c.Wait()
}

// Wait holds the main program thread until shutdown signal is received
func (c *Core) Wait() {
	sig := make(chan os.Signal, 1)

	if c.mode == "managed" {
		c.vi("managed mode; listening for sigusr2; ignoring sigusr1, sigint")
		signal.Notify(sig, syscall.SIGUSR2)
		signal.Ignore(syscall.SIGUSR1, syscall.SIGINT)
	} else {
		signal.Notify(sig, syscall.SIGINT)
	}

	c.v("main thread waiting")
	_ = <-sig
	fmt.Println() //dead line to line up output

	c.shutdown(false, "signal")
	return
}

// shutdown begins graceful shutdown procedures
func (c *Core) shutdown(remote bool, source string) {
	c.v("shutdown procedure started from " + source)

	// send shutdown notification to all services
	// if c.mode != "managed" {
	done := make(chan bool)
	fmt.Println("sending shutdown notice")
	go c.Router.sendShutdownNotices(done)
	fmt.Println("waiting")
	<-done
	fmt.Println("done")
	// time.Sleep(time.Second * 3)
	// }

	return
}

// v verbose helper
func (c *Core) v(format string, a ...interface{}) {
	notify.LnCyanF("[core] "+format, a...)
}

// ve verbose helper
func (c *Core) ve(format string, a ...interface{}) {
	notify.LnRedF("[core] "+format, a...)
}

// vi verbose helper
func (c *Core) vi(format string, a ...interface{}) {
	notify.LnYellowF("[core] "+format, a...)
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

/**********************************************************************************
**** Router
**********************************************************************************/

// Router handles all of the addressing and mapping of services that are attached to gmbhCore
type Router struct {

	// services (Name|Alias)->Service
	// map contains all registered services
	services map[string]*GmbhService

	// serviceNames is a list of the names of all services attached. This is useful because if the
	// map is walked using a range it will return a value for every alias and thus have duplicates
	serviceNames []string

	// idCounter keeps track of the current runnig id
	idCounter int

	// addressHandler is in charge of assigning addresses and making sure that there are no collisions
	addressing *addressHandler

	verbose bool
	mu      *sync.Mutex
}

// NewRouter instantiates and returns a new Router structure
func NewRouter() *Router {

	r := &Router{
		services:     make(map[string]*GmbhService),
		serviceNames: make([]string, 0),
		idCounter:    100,
		addressing: &addressHandler{
			host: config.Localhost,
			port: config.ServicePort,
			mu:   &sync.Mutex{},
		},
		mu:      &sync.Mutex{},
		verbose: true,
	}

	go r.pingHandler()

	return r
}

// LookupService looks through the services map and returns the service if it exists
func (r *Router) LookupService(name string) (*GmbhService, error) {
	r.v("looking up %s", name)
	retrievedService := r.services[name]
	if retrievedService == nil {
		r.v("not found")
		return nil, errors.New("router.LookupService.NotFound")
	}
	r.v("found")
	return retrievedService, nil
}

// AddService attaches a service to gmbH
func (r *Router) AddService(name string, aliases []string) (*GmbhService, error) {

	newService := NewService(
		r.assignNextID(),
		name,
		aliases,
		r.addressing.assignAddress(),
	)

	// check to see if it exists in map already
	s, err := r.LookupService(name)
	if err == nil {
		r.v("found new service already in map")
		if s.State == Shutdown {
			r.v("state is reported as shutdown")
			r.v("acting as if this is the same service")
			s.UpdateState(Running)
			return s, nil
		}
		r.v("state was not reported as shutdown, probable err")
	}

	err = r.addToMap(newService)
	if err != nil {
		r.v(newService.String())
		r.v("could not add service to map; err=%s", err.Error())
		return nil, err
	}

	r.v("added service=%s", newService.String())
	return newService, nil
}

// Verify a ping
func (r *Router) Verify(name, fp string) error {
	s := r.services[name]
	if s == nil {
		return errors.New("verify.notFound")
	}
	if r.services[name].Fingerprint != fp {
		return errors.New("verify.fingerprintMismatch")
	}
	if s.State == Shutdown {
		return errors.New("verify.reportedShutdown")
	}
	if s.State == Unresponsive {
		s.UpdateState(Running)
	}
	s.LastPing = time.Now()
	return nil
}

// addToMap returns an error if there is a name or alias conflict with an existing
// service in the service map, otherwise the service's name and alias are added to
// the map
func (r *Router) addToMap(newService *GmbhService) error {

	if _, ok := r.services[newService.Name]; ok {
		r.v("could not add to map, duplicate name")
		return errors.New("router.addToMap: duplicate service with same name found")
	}

	for _, alias := range newService.Aliases {
		if _, ok := r.services[alias]; ok {
			r.v("could not add to map, duplicate alias=" + alias)
			return errors.New("router.addToMap: duplicate service with same alias found")
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[newService.Name] = newService
	r.serviceNames = append(r.serviceNames, newService.Name)
	for _, alias := range newService.Aliases {
		r.services[alias] = newService
	}

	r.v("added %s to map", newService.Name)

	return nil
}

// sendShutdownNotices sends a notice to all clients that core is shutting down
func (r *Router) sendShutdownNotices(done chan bool) {
	var wg sync.WaitGroup
	for _, name := range r.serviceNames {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			service := r.services[n]
			r.v("sending shutdown to %s at %s", service.Name, service.Address)
			client, ctx, can, err := rpc.GetCabalRequest(service.Address, time.Millisecond*500)
			if err != nil {
				r.v("could not create client")
				can()
				return
			}
			req := &intrigue.ServiceUpdate{
				Request: "core.shutdown",
				Message: service.Name,
			}
			_, err = client.UpdateRegistration(ctx, req)
			if err != nil {
				if service.State != Shutdown {
					r.v("error contacting service; id=%s; err=%s", service.ID, err.Error())
				}
			}
		}(name)
	}
	wg.Wait()
	done <- true
}

// GetCoreServiceData queries each attached client to respond with their information using their
// fingerprint for validation
func (r *Router) GetCoreServiceData(core *intrigue.CoreService) []*intrigue.CoreService {
	ret := []*intrigue.CoreService{core}
	for _, n := range r.serviceNames {
		service := r.services[n]
		r.v("sending summary request to %s at %s", service.Name, service.Address)
		client, ctx, can, err := rpc.GetCabalRequest(service.Address, time.Second*1)
		if err != nil {
			r.v("could not create client")
			can()
			continue
		}
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"sender", "gmbhCore",
			"target", n,
			"fingerprint", service.Fingerprint,
		)
		req := &intrigue.Action{
			Request: "request.info.all",
		}
		resp, err := client.Summary(ctx, req)
		if err != nil {
			r.v("error contacting service; id=%s; err=%s", service.ID, err.Error())
			continue
		}
		if resp.GetServices() == nil {
			ret = append(ret, &intrigue.CoreService{
				Name:   n,
				Errors: []string{"could not contact, err=" + resp.GetError()},
			})
			continue
		}
		ret = append(ret, resp.Services...)
	}
	return ret
}

// pingHandler looks through each of the remotes in the map. if it has been more than n amount of
// time since a remote has sent a ping, it will be pinged. If the ping is not retured after n more
// seconds, the remote will be marked as Failed After n amount of time, failed remotes will
// be removed from the map
func (r *Router) pingHandler() {
	for {
		time.Sleep(time.Second * 180)
		for _, s := range r.serviceNames {
			if time.Since(r.services[s].LastPing) > time.Second*90 {
				r.v("marking name=%s; id=%s as Unresponsive", s, r.services[s].ID)
				r.services[s].UpdateState(Unresponsive)
			}
		}
	}
}

func (r *Router) assignNextID() string {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	r.idCounter++
	return strconv.Itoa(r.idCounter)
}

// v verbose printer
func (r *Router) v(msg string, a ...interface{}) {
	notify.LnGreenF("[rtr] "+msg, a...)
}

// addressHandler is in charge of assigning addresses to services
type addressHandler struct {
	table map[string]string
	host  string
	port  int
	mu    *sync.Mutex
}

func (a *addressHandler) assignAddress() string {
	a.setNextAddress()
	addr := a.host + ":" + strconv.Itoa(a.port)
	return addr
}

func (a *addressHandler) setNextAddress() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.port += 2
}

/**********************************************************************************
**** Service
**********************************************************************************/

// GmbhService is the data representation of a connected service
type GmbhService struct {
	// The id assigned by the router
	ID string

	// Aliases of the service
	Aliases []string

	// the name of the service
	Name string

	// the address to the service
	Address string

	// The time that the service was added to the router
	Added time.Time

	// The last known state of the service
	State State

	// The last time a ping was received
	LastPing time.Time

	// assigned by the server, the fingerprint is sent with each ping to verify id
	Fingerprint string

	mu *sync.Mutex
}

func (g *GmbhService) String() string {
	return fmt.Sprintf("name=%s; id=%s; address=%s;", g.Name, g.ID, g.Address)
}

// NewService returns a gmbhService object with data filled in
func NewService(id string, name string, aliases []string, address string) *GmbhService {
	return &GmbhService{
		ID:          id,
		Name:        name,
		Aliases:     aliases,
		Address:     address,
		Added:       time.Now(),
		State:       Running,
		LastPing:    time.Now().Add(time.Hour),
		Fingerprint: xid.New().String(),
		mu:          &sync.Mutex{},
	}
}

// UpdateState of the current state of the service
func (g *GmbhService) UpdateState(s State) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if s != g.State {
		g.v("marking %s(%s) as %s", g.Name, g.ID, s.String())
		g.State = s
	}
}

// v verbose printer
func (g *GmbhService) v(msg string, a ...interface{}) {
	notify.LnYellowF("[service] "+msg, a...)
}

// State controls the state of a remote server
type State int

const (
	// Running as normal
	Running State = 1 + iota

	// Shutdown notice received from remote
	Shutdown

	// Unresponsive if the service has not sent a ping in greater than some amount of time
	Unresponsive

	// Failed to return a pong
	Failed
)

var states = [...]string{
	"Running",
	"Shutdown",
	"Unresponsive",
	"Failed",
}

func (s State) String() string {
	if Running <= s && s <= Failed {
		return states[s-1]
	}
	return "%!State()"
}

/**********************************************************************************
**** OS Helpers
**********************************************************************************/

// getLogFile attempts to add the desired path as an extension to the current
// directory as reported by os.GetWd(). The file is then opened or created
// and returned
func getLogFile(desiredPathExt, filename string) (*os.File, error) {
	// get pwd
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// make sure that the path extension exists or make the directories needed
	dirPath := filepath.Join(dir, desiredPathExt)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	// create the file
	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
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
