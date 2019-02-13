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

	// The service to manage
	service *service.Service

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
		coreAddress: coreAddress,
		verbose:     verbose,
		mu:          &sync.Mutex{},
	}

	return r, nil
}

func (r *remote) Start() {

	// // start the service //////////////////////////////////////////////////////////

	// if config == "" {
	// 	notify.StdMsgErr("must specify config file using flags")
	// 	os.Exit(1)
	// }

	// var err error
	// r.service, err = service.NewManagedService("101", config)
	// if err != nil {
	// 	notify.StdMsgErr("could not start service; err=(" + err.Error() + ")")
	// 	os.Exit(1)
	// }

	// dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// r.service.StartLog(dir+"/gmbh", "process-manager.log")

	// pid, err := r.service.Start()
	// if err != nil {
	// 	notify.StdMsgErr("could not start service, error=(" + err.Error() + ")")
	// 	os.Exit(1)
	// } else {
	// 	notify.StdMsgGreen("started process; pid=(" + pid + ")")
	// }

	// // done starting service //////////////////////////////////////////////////////////

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
	// r.service.Kill()

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

// func startContainer() {

// 	c.mu = &sync.Mutex{}
// 	c.con = rpc.NewRemoteConnection("", &remoteServer{})
// 	c.coreAddr = "localhost:59997"
// 	c.closed = false
// 	c.to = time.Second * 5

// 	if !*c.daemon {
// 		notify.SetTag("[gmbh-pm] ")
// 		notify.StdMsg("gmbh container process manager")
// 	} else {
// 		notify.SetVerbose(false)
// 	}

// 	if *c.configPath == "" {
// 		notify.StdMsgErr("must specify a config file")
// 		os.Exit(1)
// 	}

// 	run()

// }

// func run() {
// 	var err error
// 	c.serv, err = service.NewManagedService("101", *c.configPath)
// 	if err != nil {
// 		notify.StdMsgErr("could not start service; err=(" + err.Error() + ")")
// 		os.Exit(1)
// 	}

// 	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
// 	c.serv.StartLog(dir+"/gmbh", "process-manager.log")

// 	pid, err := c.serv.Start()
// 	if err != nil {
// 		notify.StdMsgErr("could not start service, error=(" + err.Error() + ")")
// 		c.forkError = err
// 	} else {
// 		notify.StdMsgGreen("started process; pid=(" + pid + ")")
// 	}

// 	sigs := make(chan os.Signal, 1)
// 	done := make(chan bool, 1)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
// 	go func() {
// 		_ = <-sigs
// 		done <- true
// 	}()

// 	if *c.managed {
// 		go connect()
// 	}

// 	<-done
// 	fmt.Println()

// 	c.serv.Kill()

// 	if *c.managed {
// 		c.mu.Lock()
// 		c.closed = true
// 		c.mu.Unlock()

// 		disconnect()
// 	}

// 	notify.StdMsg("shutdown signal")
// 	return
// }

// /**********************************************************************************
// **** Handling Connection to core
// **********************************************************************************/

// func connect() {
// 	notify.StdMsg("connecting to gmbh-core")

// 	addr, status := makeConnectRequest()
// 	for status != nil {
// 		if status.Error() != "makeConnectRequest.fail" {
// 			notify.StdMsg("gmbh internal error")
// 			return
// 		}

// 		if c.closed {
// 			return
// 		}

// 		notify.StdMsg("could not connect; retry=(" + c.to.String() + ")")
// 		time.Sleep(c.to)
// 		addr, status = makeConnectRequest()
// 	}

// 	if addr == "" {
// 		notify.StdMsg("gmbh internal error, no address returned from core")
// 		return
// 	}

// 	c.con.SetAddress(addr)
// 	c.con.Remote = &remoteServer{}
// 	err := c.con.Connect()
// 	if err != nil {
// 		notify.StdMsgErr("gmbh connection error=(" + err.Error() + ")")
// 		return
// 	}

// 	// start a goroutine that will keep send the keep alive
// 	go sendPing()
// 	notify.StdMsgGreen("connected; address=(" + addr + ")")

// }

// // sendPing is meant to run in its own thread. It will continue to call itself or
// // return and changed the state of the connection if there is a failure reaching
// // the control server that is ran by gmbhCore
// func sendPing() {
// 	if c.con.IsConnected() {
// 		time.Sleep(time.Second * 45)
// 		notify.StdMsgBlue("-> ping")

// 		client, ctx, can, err := rpc.GetControlRequest(c.coreAddr, time.Second*5)
// 		if err != nil {
// 			notify.StdMsgErr(err.Error())
// 		}

// 		_, err = client.Alive(ctx, &cabal.Ping{Time: time.Now().Format(time.Stamp)})
// 		if err == nil {
// 			can()
// 			notify.StdMsgBlue("<- pong")
// 			sendPing()
// 		} else {
// 			failed()
// 			return
// 		}
// 	}
// 	return
// }

// func disconnect() {
// 	notify.StdMsg("disconnected")
// 	c.con.Disconnect()
// 	c.con.Server = nil

// 	if !c.closed {
// 		time.Sleep(c.to)
// 	}
// }

// func failed() {
// 	notify.StdMsg("failed to receive pong; disconnecting")
// 	c.con.Disconnect()
// 	c.con.Server = nil

// 	if !c.closed {
// 		time.Sleep(c.to)
// 		connect()
// 	}
// }

// func makeConnectRequest() (string, error) {
// 	client, ctx, can, err := rpc.GetControlRequest(c.coreAddr, time.Second*5)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer can()

// 	request := &cabal.ServiceUpdate{
// 		Sender:  "gmbh-container",
// 		Target:  "core",
// 		Message: c.serv.Static.Name,
// 		Action:  "container.register",
// 	}

// 	reply, err := client.UpdateServiceRegistration(ctx, request)
// 	if err != nil {
// 		notify.StdMsgErr("updateServiceRegistration err=(" + err.Error() + ")")
// 		return "", errors.New("makeConnectRequest.fail")
// 	}

// 	c.id = reply.GetStatus()

// 	return reply.GetAction(), nil
// }

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

		if !c.closed {
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
		response := &cabal.Action{
			// Sender:      r.service.Static.Name,
			Sender:  "name",
			Target:  "gmbh-core",
			Message: "response.info",
			// ServiceInfo: serviceToRPC(r.service),
		}
		return response, nil
	} else if in.GetAction() == "service.restart" {

		response := &cabal.Action{
			// Sender:  r.service.Static.Name,
			Sender:  "name",
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
