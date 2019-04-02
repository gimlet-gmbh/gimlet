package remote

/*
 * main.go (gmbhContainer)
 * Abe Dick
 * February 2019
 */

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/address"
	"github.com/gmbh-micro/rpc/intrigue"
	"github.com/gmbh-micro/service"
)

// Remote ; as in remote process manager client; holds the process
type Remote struct {

	// the entry point for services to manage
	serviceManager *ServiceManager

	// registration with data from core
	reg *registration

	// the id as assigned by core
	id string

	// the connection handler to gmbh over control server
	con *rpc.Connection

	startTime time.Time
	logPath   string
	errors    []error

	// The mode as read by the environment
	env string

	addr  *address.Handler
	raddr string

	// in containers, this will be the hostname of the container currently being
	// executed inside of
	host string

	// gmbhShutdown is marked true when sigusr1 has been sent to the pm process.
	// When this flag is true services that exit will not be restarted as it is
	// to be used in conjunction with the gmbh process launcher for graceful
	// shutdown
	gmbhShutdown bool

	// closed is set true when shutdown procedures have been started
	// either remotely or signaled from core
	closed bool

	// should the client run in verbose mode
	verbose bool

	// coreAddress is the address to core
	coreAddress string

	mu *sync.Mutex
}

// registration contains data that is received from core at registration time
type registration struct {
	// id from core
	id string

	// mode from core
	mode string

	// address to run internal server on
	address string

	// filesystem path back to core
	corePath string

	// an id given from core to send ping requests with
	fingerprint string
}

/**********************************************************************************
**** Client Operations
**********************************************************************************/

var r *Remote

// NewRemote returns a new remote object
func NewRemote(procmAddr, env string, verbose bool) (*Remote, error) {

	if r != nil {
		return r, nil
	}

	r = &Remote{
		serviceManager: NewServiceManager(),
		startTime:      time.Now(),
		coreAddress:    procmAddr,
		verbose:        verbose,
		env:            env,
		errors:         make([]error, 0),
		mu:             &sync.Mutex{},
	}

	if env == "C" {
		r.host = os.Getenv("HOSTNAME")
		r.addr = address.NewHandler(r.host, config.RemotePort+2, config.RemotePort+1002)
	} else {
		r.host = config.Localhost
	}

	return r, nil
}

// Start the remote
func (r *Remote) Start() {

	println("                    _                       ")
	println("  _  ._ _  |_  |_| |_)  _  ._ _   _ _|_  _  ")
	println(" (_| | | | |_) | | | \\ (/_ | | | (_) |_ (/_ ")
	println("  _|                                          ")
	print("started, time=" + time.Now().Format(time.Stamp))
	print("env=%s; ProcmAddress=%s; hostname=%s", r.env, r.coreAddress, r.host)

	// setting mode and choosing shutdown mechanism
	sig := make(chan os.Signal, 1)
	if r.env == "M" {
		print("remote is in managed mode; using sigusr2; ignoring sigusr1, sigint")
		signal.Notify(sig, syscall.SIGUSR2)
		signal.Ignore(syscall.SIGINT, syscall.SIGUSR1)
	} else {
		print("remote is in standalone mode; using sigint")
		signal.Notify(sig, syscall.SIGINT)
	}

	r.startTime = time.Now()
	go r.connect()

	_ = <-sig
	fmt.Println()
	r.shutdown("signal")
}

// shutdown procedures
func (r *Remote) shutdown(src string) {
	print("[remote] Shutdown procedures started in remote from " + src)
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()

	r.disconnect()

	// shutdown service
	r.serviceManager.Shutdown()
	r.notifyCore()

	print("[remote] shutdown, time=" + time.Now().Format(time.Stamp))
	os.Exit(0)
}

// connect to core if not already connected. If the error returned from making the requst
// is that core is unavailable, set a try again for every n seconds. Otherwise start a ping
// and pong response to keep track of the connection.
func (r *Remote) connect() {

	// when failed or disconnected, the registration is wiped to make sure that
	// legacy data does not get used, thus if g.reg is not nil, then we can assume
	// that a thread has aready requested and received a valid registration
	// and the current thread can be closed
	if r.reg != nil {
		return
	}

	print("attempting to connect to core @ " + r.coreAddress)

	reg, status := r.makeCoreConnectRequest()
	for status != nil {
		if status.Error() != "registration.Unavailable" {
			perr("internal error=" + status.Error())
			return
		}

		if r.closed || (r.con != nil && r.con.IsConnected()) {
			return
		}

		perr("Could not reach core, try again in 5s")
		time.Sleep(time.Second * 5)
		reg, status = r.makeCoreConnectRequest()
	}

	if reg.address == "" {
		return
	}

	r.mu.Lock()
	r.reg = reg
	r.id = reg.id
	r.con = rpc.NewRemoteConnection(reg.address, &remoteServer{})
	r.mu.Unlock()

	err := r.con.Connect()
	if err != nil {
		perr("connection error=" + err.Error())
		perr("handle this; for now return")
		r.closed = true
		return
	}
	print("connected")
}

func (r *Remote) disconnect() {
	print("disconnecting")
	r.mu.Lock()
	if r.con != nil {
		r.con.Disconnect()
	}
	r.con = nil
	r.reg = nil
	r.mu.Unlock()
}

func (r *Remote) failed() {
	print("connection to core reporting failure")
	if r.con.IsConnected() {
		r.con.Disconnect()
	}
	r.mu.Lock()
	r.con = nil
	r.mu.Unlock()

	if !r.closed {
		time.Sleep(time.Second * 5)
		print("attempting to reconneced")
		r.mu.Lock()
		r.reg = nil
		r.mu.Unlock()
		r.connect()
	}

}

func (r *Remote) makeCoreConnectRequest() (*registration, error) {
	client, ctx, can, err := rpc.GetControlRequest(r.coreAddress, time.Second*10)
	if err != nil {
		return nil, errors.New("registration.Unavailable")
	}
	defer can()

	// env = "C" - must assign own addresses
	addr := ""
	if r.env == "C" {
		if r.raddr == "" {
			addr, _ = r.addr.NextAddress()
		} else {
			addr = r.raddr
		}
		print("remote address=%s", addr)
	}

	request := &intrigue.ServiceUpdate{
		Request: "remote.register",
		Message: r.env,
		Env:     r.env,
		Address: addr,
	}

	reply, err := client.UpdateRegistration(ctx, request)
	if err != nil {
		return nil, errors.New("registration.Unavailable")
	}

	if reply.GetMessage() != "registered" {
		return nil, errors.New(reply.GetError())
	}

	reg := &registration{
		id:          reply.GetServiceInfo().GetID(),
		address:     reply.GetServiceInfo().GetAddress(),
		fingerprint: reply.GetServiceInfo().GetFingerprint(),
	}

	if r.env == "C" {
		reg.address = addr
	}

	print("registration; id=" + reg.id + "; address=" + reg.address + "; fingerprint=" + reg.fingerprint)

	return reg, nil
}

// AddService attaches services to the remote and then attempts to start them
// EXPERIMENTAL -- in a goroutine wait until the registration has been returned from
//                 procm before starting the service
func (r *Remote) AddService(conf *config.ServiceConfig) {
	go func() {
		for {
			time.Sleep(time.Second * 3)
			if r.id != "" {
				service, err := r.serviceManager.AddServiceFromConfig(conf)
				if err != nil {
					perr("could add start service; error=" + err.Error())
					return
				}
				service.Static.Env = append(service.Static.Env, "REMOTE="+r.id)
				pid, err := service.Start(r.env, r.verbose)
				if err != nil {
					perr("could not start service; error=" + err.Error())
					return
				}
				print("service started with pid=%s", pid)
				break
			}
		}
	}()
}

// GetServices returns all service pointers attached to the Remote
func (r *Remote) GetServices() []*service.Service {
	return r.serviceManager.GetAllServices()
}

// RestartAll services attached
func (r *Remote) RestartAll() {
	r.serviceManager.RestartAll()
}

// Restart service with id
func (r *Remote) Restart(id string) (pid string, err error) {
	return r.serviceManager.Restart(id)
}

// LookupService returns the service with the id or returns error
func (r *Remote) LookupService(id string) (*service.Service, error) {
	service, err := r.serviceManager.LookupByID(id)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// notifyCore of shutdown
func (r *Remote) notifyCore() {
	print("sending notify to core")
	if r.id == "" {
		print("invalid id")
		return
	}

	client, ctx, can, err := rpc.GetControlRequest(r.coreAddress, time.Second)
	if err != nil {
		return
	}
	defer can()

	request := &intrigue.ServiceUpdate{
		Message: r.id,
		Request: "shutdown.notif",
	}
	client.UpdateRegistration(ctx, request)
	print("notice sent")
	return
}

// GetRegistration returns the processes registration or an empty reg if nil
func (r *Remote) GetRegistration() *registration {
	if r.reg != nil {
		return r.reg
	}
	return &registration{}
}

/**********************************************************************************
**** RPC server
**********************************************************************************/

type remoteServer struct{}

func (s *remoteServer) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {
	// md, ok := metadata.FromIncomingContext(ctx)

	vrpc("-> %s", in.String())

	request := in.GetRequest()

	if request == "core.shutdown" {

		if r.env == "M" {
			go r.shutdown("procm")
		} else if !r.closed {
			go func() {
				r.disconnect()
				r.connect()
			}()
		}

		return &intrigue.Receipt{
			Message: "ack",
		}, nil

	} else if request == "gmbh.shutdown" {
		r.mu.Lock()
		r.gmbhShutdown = true
		r.mu.Unlock()
		go r.serviceManager.NotifyGracefulShutdown()

		return &intrigue.Receipt{
			Message: "ack",
		}, nil

	}

	return &intrigue.Receipt{Error: "requst.unknown"}, nil
}

func (s *remoteServer) NotifyAction(ctx context.Context, in *intrigue.Action) (*intrigue.Action, error) {
	// md, ok := metadata.FromIncomingContext(ctx)

	vrpc("-> %s", in.String())

	request := in.GetRequest()
	TargetID := in.GetTarget()

	if request == "service.restart.one" {

		service, err := r.LookupService(TargetID)
		if err != nil {
			return &intrigue.Action{Error: "service.notFound"}, nil
		}

		pid, err := service.Restart()
		if err != nil {
			return &intrigue.Action{Error: "service.restartError=" + err.Error()}, nil
		}

		return &intrigue.Action{
			Message: pid,
		}, nil

	} else if request == "service.restart.all" {
		go r.RestartAll()
		return &intrigue.Action{Message: "success"}, nil
	}

	return &intrigue.Action{Error: "request.unknown"}, nil
}

func (s *remoteServer) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {
	// md, ok := metadata.FromIncomingContext(ctx)

	vrpc("-> %s", in.String())

	request := in.GetRequest()
	targetID := in.GetTarget()

	if request == "request.info.all" {

		services := r.GetServices()
		rpcServices := []*intrigue.Service{}
		for _, service := range services {
			rpcServices = append(rpcServices, serviceToRPC(service))
		}

		errs := []string{}
		for _, e := range r.errors {
			errs = append(errs, e.Error())
		}
		stat := "Stable"
		if len(errs) != 0 {
			stat = "Degraded"
		}

		return &intrigue.SummaryReceipt{
			Remotes: []*intrigue.ProcessManager{
				&intrigue.ProcessManager{
					ID:        r.id,
					Address:   r.GetRegistration().address,
					StartTime: r.startTime.Format(time.RFC3339),
					Errors:    errs,
					Status:    stat,
					LogPath:   r.logPath,
					Services:  rpcServices,
				},
			},
		}, nil

	} else if request == "request.info.one" {

		service, err := r.LookupService(targetID)
		if err != nil {
			vrpc("not found")
			return &intrigue.SummaryReceipt{Error: "service.notFound"}, nil
		}

		vrpc("returning service info " + service.ID)

		errs := []string{}
		for _, e := range r.errors {
			errs = append(errs, e.Error())
		}
		stat := "Stable"
		if len(errs) != 0 {
			stat = "Degraded"
		}

		return &intrigue.SummaryReceipt{
			Remotes: []*intrigue.ProcessManager{
				&intrigue.ProcessManager{
					ID:        r.id,
					Address:   r.GetRegistration().address,
					StartTime: r.startTime.Format(time.RFC3339),
					Errors:    errs,
					Status:    stat,
					LogPath:   r.logPath,
					Services:  []*intrigue.Service{serviceToRPC(service)},
				},
			},
		}, nil

	}

	return &intrigue.SummaryReceipt{Error: "request.unknown"}, nil
}

func (s *remoteServer) Alive(ctx context.Context, in *intrigue.Ping) (*intrigue.Pong, error) {
	return &intrigue.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func vrpc(format string, a ...interface{}) {
	print("[rpc] "+format, a...)
}

func serviceToRPC(s *service.Service) *intrigue.Service {

	procRuntime := s.Process.GetInfo()

	si := &intrigue.Service{
		Id: r.id + "-" + s.ID,
		// Name:      s.Static.Name,
		Status:    s.Process.GetStatus().String(),
		Path:      "-",
		LogPath:   s.LogPath,
		Pid:       int32(procRuntime.PID),
		Fails:     int32(procRuntime.Fails),
		Restarts:  int32(procRuntime.Restarts),
		StartTime: procRuntime.StartTime.Format(time.RFC3339),
		FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
		Errors:    s.Process.GetErrors(),
		Mode:      s.Mode.String(),
	}

	return si
}

/**********************************************************************************
**** Service Manager
**********************************************************************************/

// ServiceManager controls all of the attachable services to the remote process manager
type ServiceManager struct {

	// services listed from map[id]*service
	services map[string]*service.Service

	idCounter int

	mu *sync.Mutex
}

// NewServiceManager instantiates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services:  make(map[string]*service.Service),
		idCounter: 100,
		mu:        &sync.Mutex{},
	}
}

// AddServiceFromConfig attaches a service to the service manager by parsing the config file at configPath
// and creating a new local binary manager from the process package
func (s *ServiceManager) AddServiceFromConfig(conf *config.ServiceConfig) (*service.Service, error) {

	newService, err := service.NewService(s.assignID(), conf)
	if err != nil {
		return nil, errors.New("serviceManager.AddServiceFromConfig.serviceErr=" + err.Error())
	}

	err = s.addToMap(newService)
	if err != nil {
		return nil, errors.New("serviceManager.AddServiceFromConfig.serviceErr=" + err.Error())
	}

	print("added " + newService.ID)

	return newService, nil
}

// NotifyGracefulShutdown of all attached services
func (s *ServiceManager) NotifyGracefulShutdown() {
	print("sending graceful shutdown notices")
	for _, v := range s.services {
		v.EnableGracefulShutdown()
	}
}

// GetAllServices returns the contents of the service map
func (s *ServiceManager) GetAllServices() []*service.Service {
	ret := []*service.Service{}
	for _, v := range s.services {
		ret = append(ret, v)
	}
	return ret
}

// LookupByID returns the service with id or else error
func (s *ServiceManager) LookupByID(id string) (*service.Service, error) {
	service := s.services[id]
	if service == nil {
		return nil, errors.New("serviceManager.lookup.notFound")
	}
	return service, nil
}

// Shutdown kills all attached processes
func (s *ServiceManager) Shutdown() {
	for _, s := range s.services {
		print("sending shutdown to " + s.ID)
		s.Kill()
	}
}

// RestartAll attached processes
func (s *ServiceManager) RestartAll() {
	for _, s := range s.services {
		print("sending restart to " + s.ID)
		pid, err := s.Restart()
		if err != nil {
			perr("could not restart; err=" + err.Error())
		}
		print("Pid=" + pid)
	}
}

// Restart attached processe with id
func (s *ServiceManager) Restart(id string) (pid string, err error) {
	service := s.services[id]
	if service == nil {
		return "-1", errors.New("serviceManager.Restart.NotFound")
	}
	return service.Restart()
}

// addToMap adds the service to the map or returns error
func (s *ServiceManager) addToMap(newService *service.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[newService.ID]; ok {
		return errors.New("serviceManager.addToMap.error")
	}
	s.services[newService.ID] = newService
	return nil
}

// returns the next id string and then increments the id counter
func (s *ServiceManager) assignID() string {

	s.mu.Lock()
	defer s.mu.Unlock()
	id := "s" + strconv.Itoa(s.idCounter)
	s.idCounter++
	return id

}

func print(format string, a ...interface{}) {
	tag := "[" + r.id + "] "
	if r.id == "" {
		tag = "[remote] "
	}
	notify.LnMagentaF(tag+format, a...)
}

func println(format string, a ...interface{}) {
	notify.LnMagentaF(format, a...)
}

func perr(format string, a ...interface{}) {
	tag := "[" + r.id + "] "
	if r.id == "" {
		tag = "[remote] "
	}
	notify.LnRedF(tag+format, a...)
}
