package main

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

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/process"
)

type remote struct {

	// the entry point for services to manage
	serviceManager *ServiceManager

	// registration with data from core
	reg *registration

	// the connection handler to gmbh over control server
	con *rpc.Connection

	// ping helpers help with gothread synchronization
	pingHelpers []*pingHelper

	pingCounter int
	startTime   time.Time

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
}

/**********************************************************************************
**** Client Operations
**********************************************************************************/

var r *remote

func newRemote(coreAddress string, verbose bool) (*remote, error) {

	if r != nil {
		return r, nil
	}

	r = &remote{
		serviceManager: NewServiceManager(),
		pingHelpers:    make([]*pingHelper, 0),
		startTime:      time.Now(),
		coreAddress:    coreAddress,
		verbose:        verbose,
		mu:             &sync.Mutex{},
	}

	return r, nil
}

func (r *remote) Start() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()

	notify.StdMsgNoPrompt("------------------------------------------------------------")
	notify.StdMsg("started, time=" + time.Now().Format(time.Stamp))
	r.startTime = time.Now()

	go r.connect()

	<-done
	fmt.Println()

	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()

	r.disconnect()

	// shutdown service
	r.serviceManager.Shutdown()

	// todo: send message to core of shutdown

	notify.StdMsgBlue("shutdown, time=" + time.Now().Format(time.Stamp))

	p := int64(time.Since(r.startTime) / (time.Second * 45))

	notify.StdMsgBlue("Ping counter should be around " + strconv.Itoa(int(p)))
	os.Exit(0)
}

// connect to core if not already connected. If the error returned from making the requst
// is that core is unavailable, set a try again for every n seconds. Otherwise start a ping
// and pong response to keep track of the connection.
func (r *remote) connect() {

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

	notify.StdMsgBlue("registration details")
	notify.StdMsgBlue("id=" + reg.id + "; address=" + reg.address)

	if reg.address == "" {
		return
	}

	r.mu.Lock()
	r.reg = reg
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

func (r *remote) disconnect() {
	notify.StdMsgBlue("disconnecting")
	r.mu.Lock()
	if r.con != nil {
		r.con.Disconnect()
	}
	r.con = nil
	r.reg = nil
	r.mu.Unlock()
}

func (r *remote) failed() {
	notify.StdMsgBlue("connection to core reporting failure")
	if r.con.IsConnected() {
		r.con.Disconnect()
	}
	r.mu.Lock()
	r.con = nil
	r.mu.Unlock()

	if !r.closed {
		time.Sleep(time.Second * 5)
		r.connect()
	}

}

func (r *remote) sendPing(ph *pingHelper) {
	for {

		time.Sleep(time.Second * 45)

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

			_, err = client.Alive(ctx, &cabal.Ping{Time: time.Now().Format(time.Stamp)})
			can()
			if err == nil {
				notify.StdMsgBlue("<- pong")
			} else {
				r.failed()
				return
			}
		}
	}
}

func (r *remote) makeCoreConnectRequest() (*registration, error) {
	client, ctx, can, err := rpc.GetControlRequest(r.coreAddress, time.Second*10)
	if err != nil {
		// panic(err)
		return nil, errors.New("registration.Unavailable")
	}
	defer can()

	request := &cabal.ServiceUpdate{
		Sender:  "gmbh-remote",
		Target:  "core",
		Message: "announce",
		Action:  "remote.register",
	}

	reply, err := client.UpdateServiceRegistration(ctx, request)
	if err != nil {
		// notify.StdMsgErr("updateServiceRegistration err=(" + err.Error() + ")")
		return nil, errors.New("registration.Unavailable")
	}

	fmt.Println("reply is")
	fmt.Println(*reply)

	if reply.GetMessage() != "registered" {
		return nil, errors.New(reply.GetMessage())
	}

	reg := &registration{
		id:      reply.GetTarget(),
		address: reply.GetStatus(),
	}

	return reg, nil
}

// AddService attaches services to the remote and then attempts to start them
func (r *remote) AddService(configPath string) (pid string, err error) {
	service, err := r.serviceManager.AddServiceFromConfig(configPath)
	if err != nil {
		return "-1", errors.New("could not start service; error=" + err.Error())
	}
	return service.Start()

}

// GetService returns all service pointers attached to the remote
func (r *remote) GetServices() []*service.Service {
	return r.serviceManager.GetAllServices()
}

/**********************************************************************************
**** REFACTORED ABOVE THIS LINE
**********************************************************************************/

type container struct {
	serv     *service.Service
	reg      *registration
	con      *rpc.Connection
	to       time.Duration
	mu       *sync.Mutex
	coreAddr string

	// closed is set true when shutdown procedures have been started
	closed bool

	id        string
	forkError error

	configPath *string
	managed    *bool
	embedded   *bool
	daemon     *bool
}

/**********************************************************************************
**** RPC server
**********************************************************************************/

type remoteServer struct{}

func (s *remoteServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
	notify.StdMsgBlue(fmt.Sprintf("-> Update Service Request; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetAction() == "core.shutdown" {
		response := &cabal.ServiceUpdate{
			Sender:  r.reg.id,
			Target:  "gmbh-core",
			Message: "ack",
		}

		r.pingHelpers = broadcast(r.pingHelpers)

		if !r.closed {
			go func() {
				r.disconnect()
				r.connect()
			}()
		}
		return response, nil
	}

	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

func (s *remoteServer) RequestRemoteAction(ctx context.Context, in *cabal.Action) (*cabal.Action, error) {
	notify.StdMsgBlue(fmt.Sprintf("-> Request Remote Action; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetAction() == "request.info" {

		services := r.GetServices()
		rpcServices := []*cabal.Service{}
		for _, service := range services {
			rpcService := &cabal.Service{
				Id:   service.ID,
				Name: service.Static.Name,
			}
			rpcServices = append(rpcServices, rpcService)
		}
		response := &cabal.Action{
			Sender:   r.reg.id,
			Target:   "gmbh-core",
			Message:  "response.info",
			Services: rpcServices,
		}
		return response, nil
	} else if in.GetAction() == "service.restart" {

		response := &cabal.Action{
			Sender:  r.reg.id,
			Target:  "gmbh-core",
			Message: "action.completed",
		}

		// pid, err := r.service.Restart()
		// if err != nil {
		// 	response.Status = err.Error()
		// } else {
		// 	response.Status = pid
		// }

		notify.StdMsgBlue(fmt.Sprintf("<- Message=(%s); Status=(%s)", response.Message, response.Status))
		return response, nil
	}
	return &cabal.Action{Message: "unimp"}, nil

}

func (s *remoteServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func serviceToRPC(s *service.Service) *cabal.Service {
	procRuntime := c.serv.Process.GetInfo()

	si := &cabal.Service{
		Id:        c.serv.ID,
		Name:      c.serv.Static.Name,
		Path:      "-",
		LogPath:   "-",
		Pid:       0,
		Fails:     int32(procRuntime.Fails),
		Restarts:  int32(procRuntime.Restarts),
		StartTime: procRuntime.StartTime.Format(time.RFC3339),
		FailTime:  procRuntime.DeathTime.Format(time.RFC3339),
		Errors:    c.serv.Process.GetErrors(),
		Mode:      "remote",
	}

	switch c.serv.Process.GetStatus() {
	case process.Stable:
		si.Status = "Stable"
	case process.Running:
		si.Status = "Running"
	case process.Failed:
		si.Status = "Failed"
	case process.Killed:
		si.Status = "Killed"
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

// GetAllServices returns the contents of the service map
func (s *ServiceManager) GetAllServices() []*service.Service {
	ret := []*service.Service{}
	for _, v := range s.services {
		ret = append(ret, v)
	}
	return ret
}

// Shutdown kills all attached processes
func (s *ServiceManager) Shutdown() {
	for _, s := range s.services {
		notify.StdMsgBlue("sending shutdown to " + s.ID)
		s.Kill()
	}
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
