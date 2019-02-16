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

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"github.com/gmbh-micro/service"
	"google.golang.org/grpc/metadata"
)

type Remote struct {

	// the entry point for services to manage
	serviceManager *ServiceManager

	// registration with data from core
	reg *registration

	// the id as assigned by core
	id string

	// the connection handler to gmbh over control server
	con *rpc.Connection

	// ping helpers help with gothread synchronization
	pingHelpers []*pingHelper

	pingCounter int
	PongDelay   time.Duration
	startTime   time.Time

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

func NewRemote(coreAddress string, verbose bool) (*Remote, error) {

	if r != nil {
		return r, nil
	}

	r = &Remote{
		serviceManager: NewServiceManager(),
		pingHelpers:    make([]*pingHelper, 0),
		PongDelay:      time.Second * 45,
		startTime:      time.Now(),
		coreAddress:    coreAddress,
		verbose:        verbose,
		mu:             &sync.Mutex{},
	}

	if verbose {
		notify.LnBYellowF("                      _                       ")
		notify.LnBYellowF("  _  ._ _  |_  |_|   |_)  _  ._ _   _ _|_  _  ")
		notify.LnBYellowF(" (_| | | | |_) | |   | \\ (/_ | | | (_) |_ (/_ ")
		notify.LnBYellowF("  _|                                          ")
	}
	return r, nil
}

func (r *Remote) Start() {

	sig := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	if os.Getenv("PMMODE") == "PMManaged" {
		notify.StdMsgBlue("overriding sigusr2")
		notify.StdMsgBlue("ignoring sigint, sigusr1")
		signal.Notify(sig, syscall.SIGUSR2)
		signal.Ignore(syscall.SIGINT, syscall.SIGUSR1)
	} else {
		notify.StdMsgBlue("overriding sigint")
		signal.Notify(sig, syscall.SIGINT)
	}
	go func() {
		_ = <-sig
		done <- true
	}()

	notify.StdMsgNoPrompt("------------------------------------------------------------")
	notify.StdMsg("started, time=" + time.Now().Format(time.Stamp))
	r.startTime = time.Now()

	go r.connect()

	<-done
	fmt.Println()

	r.shutdown()

}

// shutdown procedures
func (r *Remote) shutdown() {
	notify.LnBYellowF("Shutdown procedures started in remote")
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()

	r.disconnect()

	// shutdown service
	r.serviceManager.Shutdown()

	r.notifyCore()

	notify.LnBYellowF("shutdown, time=" + time.Now().Format(time.Stamp))

	p := int64(time.Since(r.startTime) / (time.Second * 45))

	notify.LnBYellowF("Ping counter should be around " + strconv.Itoa(int(p)))
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

	notify.StdMsgBlue("attempting to connect to core")

	reg, status := r.makeCoreConnectRequest()
	for status != nil {
		if status.Error() != "registration.Unavailable" {
			notify.StdMsgErr("internal error=" + status.Error())
			return
		}

		if r.closed || (r.con != nil && r.con.IsConnected()) {
			return
		}

		notify.StdMsgErr("Could not reach core, try again in 5s")
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
	ph := newPingHelper()
	r.pingHelpers = append(r.pingHelpers, ph)
	r.mu.Unlock()

	err := r.con.Connect()
	if err != nil {
		notify.StdMsgErr("connection error=" + err.Error())
		notify.StdMsgErr("handle this; for now return")
		r.closed = true
		return
	}
	notify.StdMsgBlue("connected")

	go r.sendPing(ph)
}

func (r *Remote) disconnect() {
	notify.StdMsgBlue("disconnecting")
	r.mu.Lock()
	if r.con != nil {
		r.con.Disconnect()
	}
	r.con = nil
	r.reg = nil
	r.mu.Unlock()
}

func (r *Remote) failed() {
	notify.StdMsgBlue("connection to core reporting failure")
	if r.con.IsConnected() {
		r.con.Disconnect()
	}
	r.mu.Lock()
	r.con = nil
	r.mu.Unlock()

	if !r.closed {
		time.Sleep(time.Second * 5)
		notify.StdMsgBlue("attempting to reconneced")
		r.mu.Lock()
		r.reg = nil
		r.mu.Unlock()
		r.connect()
	}

}

func (r *Remote) sendPing(ph *pingHelper) {
	for {

		time.Sleep(r.PongDelay)

		r.mu.Lock()
		r.pingCounter++
		r.mu.Unlock()

		notify.StdMsgBlue("-> ping " + strconv.Itoa(r.pingCounter))

		select {
		case _ = <-ph.pingChan: // case in which this channel has a message in the buffer
			close(ph.pingChan)
			notify.StdMsgBlue("<- buffer")
			ph.mu.Lock()
			ph.received = true
			ph.mu.Unlock()
			return
		default:
			if r.con == nil || (r.con != nil && !r.con.IsConnected()) {
				return
			}

			client, ctx, can, err := rpc.GetControlRequest(r.coreAddress, time.Second*30)
			if err != nil {
				notify.StdMsgErr(err.Error())
				r.failed()
			}

			ctx = metadata.AppendToOutgoingContext(
				ctx,
				"sender", r.id,
				"target", "procm",
				"fingerprint", r.reg.fingerprint,
			)

			pong, err := client.Alive(ctx, &intrigue.Ping{Status: r.id, Time: time.Now().Format(time.Stamp)})
			can()
			if err != nil {
				notify.StdMsgErr("did not receive pong response")
				r.failed()
				return
			}
			if pong.GetError() == "" {
				notify.StdMsgBlue("<- pong")
			} else {
				notify.StdMsgBlue("<- pong error=" + pong.GetError())
				r.failed()
				return
			}
		}
	}
}

func (r *Remote) makeCoreConnectRequest() (*registration, error) {
	client, ctx, can, err := rpc.GetControlRequest(r.coreAddress, time.Second*10)
	if err != nil {
		return nil, errors.New("registration.Unavailable")
	}
	defer can()

	request := &intrigue.ServiceUpdate{
		Request: "remote.register",
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

	notify.StdMsgBlue("registration; id=" + reg.id + "; address=" + reg.address + "; fingerprint=" + reg.fingerprint)

	return reg, nil
}

// AddService attaches services to the remote and then attempts to start them
func (r *Remote) AddService(configPath string) (pid string, err error) {
	service, err := r.serviceManager.AddServiceFromConfig(configPath)
	if err != nil {
		return "-1", errors.New("could not start service; error=" + err.Error())
	}

	if os.Getenv("PMMODE") == "PMManaged" {
		return service.Start("PMManaged")
	}
	return service.Start("")
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
	notify.StdMsgBlue("sending notify to core")
	if r.id == "" {
		notify.StdMsgBlue("invalid id")
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
	notify.StdMsgBlue("notice sent")
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

		r.pingHelpers = broadcast(r.pingHelpers)

		if os.Getenv("PMMODE") == "PMManaged" {
			go r.shutdown()
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

		return &intrigue.SummaryReceipt{
			Remotes: []*intrigue.ProcessManager{
				&intrigue.ProcessManager{
					ID:       r.id,
					Address:  r.GetRegistration().address,
					Services: rpcServices,
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

		return &intrigue.SummaryReceipt{
			Remotes: []*intrigue.ProcessManager{
				&intrigue.ProcessManager{
					ID:       r.id,
					Address:  r.GetRegistration().address,
					Services: []*intrigue.Service{serviceToRPC(service)},
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
	notify.LnBlueF("[rpc] "+format, a...)
}

func serviceToRPC(s *service.Service) *intrigue.Service {

	procRuntime := s.Process.GetInfo()

	si := &intrigue.Service{
		Id:        r.id + "-" + s.ID,
		Name:      s.Static.Name,
		Status:    s.Process.GetStatus().String(),
		Path:      "-",
		LogPath:   "-",
		Pid:       int32(procRuntime.PID),
		Fails:     int32(procRuntime.Fails),
		Restarts:  int32(procRuntime.Restarts),
		StartTime: procRuntime.StartTime.Format(time.RFC3339),
		FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
		Errors:    s.Process.GetErrors(),
		Mode:      "remote",
	}

	return si
}

type pingHelper struct {
	pingChan  chan bool
	contacted bool
	received  bool
	mu        *sync.Mutex
}

func newPingHelper() *pingHelper {
	return &pingHelper{
		pingChan: make(chan bool, 1),
		mu:       &sync.Mutex{},
	}
}

func broadcast(phs []*pingHelper) []*pingHelper {
	for _, p := range phs {
		p.mu.Lock()
		p.pingChan <- true
		p.contacted = true
		p.mu.Unlock()
	}
	return update(phs)
}

func update(phs []*pingHelper) []*pingHelper {
	n := []*pingHelper{}
	c := 0
	for _, p := range phs {
		if p.contacted && p.received {
			n = append(n, p)
		} else {
			c++
		}
	}
	notify.StdMsgBlue("removed " + strconv.Itoa(len(phs)-c) + "/" + strconv.Itoa(len(phs)) + " channels")
	return n
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
func (s *ServiceManager) AddServiceFromConfig(configPath string) (*service.Service, error) {

	if configPath == "" {
		return nil, errors.New("serviceManager.AddServiceFromConfig.unspecified config")
	}

	newService, err := service.NewService(s.assignID(), configPath)
	if err != nil {
		return nil, errors.New("serviceManager.AddServiceFromConfig.serviceErr=" + err.Error())
	}

	err = s.addToMap(newService)
	if err != nil {
		return nil, errors.New("serviceManager.AddServiceFromConfig.serviceErr=" + err.Error())
	}

	notify.StdMsgBlue("added " + newService.ID)

	return newService, nil
}

// NotifyGracefulShutdown of all attached services
func (s *ServiceManager) NotifyGracefulShutdown() {
	notify.StdMsgBlue("sending graceful shutdown notices")
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
		notify.StdMsgBlue("sending shutdown to " + s.ID)
		s.Kill()
	}
}

// RestartAll attached processes
func (s *ServiceManager) RestartAll() {
	for _, s := range s.services {
		notify.StdMsgBlue("sending restart to " + s.ID)
		pid, err := s.Restart()
		if err != nil {
			notify.StdMsgErr("could not restart; err=" + err.Error())
		}
		notify.StdMsgBlue("Pid=" + pid)
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
	if _, ok := s.services[newService.ID]; ok {
		return errors.New("serviceManager.addToMap.error")
	}

	s.mu.Lock()
	s.services[newService.ID] = newService
	s.mu.Unlock()

	return nil
}

// returns the next id string and then increments the id counter
func (s *ServiceManager) assignID() string {
	defer func() {
		s.mu.Lock()
		s.idCounter++
		s.mu.Unlock()
	}()
	return "s" + strconv.Itoa(s.idCounter)
}
