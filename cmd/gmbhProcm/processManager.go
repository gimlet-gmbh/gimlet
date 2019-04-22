package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/address"
	"github.com/gmbh-micro/rpc/intrigue"
	"github.com/rs/xid"
)

// ProcessManager is the main controller of the control server.
//
// It is in control of managing all remote servers. Remote servers host processes.
// It is in charge of handling all ctrl requests from the ctrl tool
type ProcessManager struct {
	Version  string
	CodeName string

	startTime time.Time

	// This is the address that will host the control server
	Address string

	// mode from env controls things such as how signals are handled
	env string

	// The connection that hosts the control server
	con *rpc.Connection

	// Router manages all addresses and instances of remotes
	router *Router

	mu         *sync.Mutex
	shutdownmu *sync.Mutex
	verbose    bool
}

var procm *ProcessManager

// NewProcessManager instantiates a new pm if one has not already been created. Note that this
// should be assigned to a global instance to interface with the rpc server. The rpc server should
// then use the GetProcM function to ensure that the global has not fallen out of scope.
func NewProcessManager(addr, env string, v bool) *ProcessManager {

	// Make sure that it is never allowed to overrite once already instantiated
	if procm != nil {
		return procm
	}

	procm = &ProcessManager{
		Version:    config.Version,
		CodeName:   config.Code,
		startTime:  time.Now(),
		Address:    addr,
		router:     NewRouter(),
		env:        env,
		verbose:    v,
		mu:         &sync.Mutex{},
		shutdownmu: &sync.Mutex{},
	}

	print("                    _                 ")
	print("  _  ._ _  |_  |_| |_) ._ _   _ |\\/| ")
	print(" (_| | | | |_) | | |   | (_) (_ |  |  ")
	print("  _|                                  ")
	print("version=%s", procm.Version)
	print("env=%s; address=%s", procm.env, procm.Address)
	return procm
}

// GetProcM is to be used by the rpc server to retrieve instances of the process manager
// asynchronously
func GetProcM() (*ProcessManager, error) {
	if procm == nil {
		return nil, errors.New("procm.nilError")
	}
	return procm, nil
}

// Start launches the grpc server using the control service in the cabal package
func (p *ProcessManager) Start() error {
	p.con = rpc.NewControlConnection(p.Address, &controlServer{})
	err := p.con.Connect()
	if err != nil {
		return err
	}
	return nil
}

// Wait sets the process manager to block the main thread until shutdown signal is recieved
// either by the terminal or using the control tool
func (p *ProcessManager) Wait() {

	go p.gracefulShutdownListener()

	// set up the listener for shutdown
	sig := make(chan os.Signal, 1)
	if p.env == "M" {
		print("procm is in managed mode; overriding sigusr2; ignoring sigint")
		signal.Notify(sig, syscall.SIGQUIT)
		signal.Ignore(syscall.SIGINT)
	} else {
		signal.Notify(sig, syscall.SIGINT)
	}

	_ = <-sig
	fmt.Println() // dead-line to align output after sigint
	p.shutdownmu.Lock()
	p.Shutdown(false)
}

// gracefulShutdownListener
func (p *ProcessManager) gracefulShutdownListener() {

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM)

	_ = <-shutdown
	// print("SIGUSR1 reported")
	p.shutdownmu.Lock()
	p.sendGmbhShutdown()
	p.shutdownmu.Unlock()
}

// RegisterRemote adds the remote to the router and sends back the id and address
func (p *ProcessManager) RegisterRemote(mode, env, addr string) (id, address, fingerprint string, err error) {
	print("registering new remote")
	rm, err := p.router.AttachNewRemote(mode, env, addr)
	if err != nil {
		return "", "", "", err
	}
	return rm.ID, rm.Address, rm.fingerprint, nil
}

// GetAllRemotes returns a reference to all attached remotes in an array
func (p *ProcessManager) GetAllRemotes() []*RemoteServer {
	return p.router.GetAllAttached()
}

// LookupRemote returns a reference to a remote if it exists in the router
func (p *ProcessManager) LookupRemote(id string) (*RemoteServer, error) {
	return p.router.LookupRemote(id)
}

// RestartRemote restarts a remote with id=id and if not found returns an error
func (p *ProcessManager) RestartRemote(parent, target string) error {
	r, err := p.router.LookupRemote(parent)
	if err != nil {
		return nil
	}
	return p.sendRestart(r.Address, r.ID, false)
}

// Verify ; as in verifyPingInfo; checks the fingerprint against the one on file, if it is a match
// it marks now as the last ping time;
func (p *ProcessManager) Verify(id, fp string) bool {
	r, e := p.LookupRemote(id)
	if e != nil {
		return false
	}
	if r.fingerprint != fp {
		return false
	}
	r.LastPing = time.Now()
	return true
}

// RestartAll sends an rpc restart request to all remotes
func (p *ProcessManager) RestartAll() []error {
	all := p.router.GetAllAttached()
	errors := []error{}
	for _, r := range all {
		err := p.sendRestart(r.Address, "", true)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		for _, e := range errors {
			print("restart error=" + e.Error())
		}
	} else {
		print("sent all restart requests with no errors")
	}
	return errors
}

// sendRestart sends a restart request to a remote
func (p *ProcessManager) sendRestart(address, id string, all bool) error {
	print("sending restart request to " + id)

	client, ctx, can, err := rpc.GetRemoteRequest(address, time.Second*2)
	if err != nil {
		return err
	}

	action := &intrigue.Action{
		Request: "service.restart.one",
		Target:  id,
	}
	if all {
		action.Request = "service.restart.all"
	}
	_, err = client.NotifyAction(ctx, action)
	if err != nil {
		return err
	}
	can()
	return nil
}

// sendGmbhShutdown sends a message to all remotes to no longer restart their attached
// processes because gmbh process manager has signaled shutdown time
func (p *ProcessManager) sendGmbhShutdown() {

	remotes := p.router.GetAllAttached()
	for _, r := range remotes {
		client, ctx, can, err := rpc.GetRemoteRequest(r.Address, time.Second*2)
		if err != nil {
			continue
		}
		update := &intrigue.ServiceUpdate{
			Request: "gmbh.shutdown",
		}
		_, err = client.UpdateRegistration(ctx, update)
		if err != nil {
			continue
		}
		can()
	}
}

// sendShutdown notice to all attached remotes
func (p *ProcessManager) sendShutdown(done chan bool) {
	var wg sync.WaitGroup
	remotes := p.router.GetAllAttached()
	for _, r := range remotes {
		wg.Add(1)
		go func(r *RemoteServer) {
			defer wg.Done()
			print("sending shutdown notice to " + r.ID)
			client, ctx, can, err := rpc.GetRemoteRequest(r.Address, time.Second*2)
			if err != nil {
				print("could not get client; err=%s", err.Error())
				return
			}
			update := &intrigue.ServiceUpdate{
				Request: "core.shutdown",
			}
			_, err = client.UpdateRegistration(ctx, update)
			if err != nil {
				print("could not contact client; err=%s", err.Error())
				return
			}
			can()
		}(r)
	}
	wg.Wait()
	done <- true
}

// MarkShutdown marks the remote as having shutdown and being inactive
func (p *ProcessManager) MarkShutdown(id string) {
	p.router.Shutdown(id)
}

// Shutdown starts shutdown procedures. If remote it indicates tat the signal came from the control
// tool
func (p *ProcessManager) Shutdown(remote bool) {
	print("shutting down...")
	noticesSent := make(chan bool)
	go p.sendShutdown(noticesSent)
	<-noticesSent
	p.con.Disconnect()

	print("shutdown; time=%s", time.Now().Format(time.Stamp))
	p.shutdownmu.Unlock()
	return
}

// Router controls the handling of attached remote servers including assigning addresses
// and managing the actual instance of each along with reporting all associated errors
type Router struct {
	// map[ID]RemoteServer holds the internal lookup of each of the attached remote servers
	remotes map[string]*RemoteServer

	// idCounter is the next id to assign
	idCounter int

	// handles the assignment of addresses for new remotes
	addr *address.Handler

	mu *sync.Mutex

	Verbose bool
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	r := &Router{
		remotes:   make(map[string]*RemoteServer),
		idCounter: 100,
		addr:      address.NewHandler(config.Localhost, config.RemotePort, config.RemotePort+1000),
		mu:        &sync.Mutex{},
		Verbose:   true,
	}
	return r
}

// LookupRemote scans through the remote map and returns if a match is found, otherwise an
// an error is returned
func (r *Router) LookupRemote(id string) (*RemoteServer, error) {
	rm := r.remotes[id]
	if rm == nil {
		print("attempted to lookkup up remote with id=" + id + "; not found")
		return nil, errors.New("router.LookupRemote.notFound")
	}
	return rm, nil
}

// AttachNewRemote adds the remote to the map and then returns the new remote server object
func (r *Router) AttachNewRemote(mode, env, addr string) (*RemoteServer, error) {

	address := addr
	if env != "C" {
		var err error
		address, err = r.addr.NextAddress()
		if err != nil {
			return nil, err
		}
	}

	newRemote := NewRemoteServer(r.assignID(), address, mode)
	err := r.addToMap(newRemote)
	if err != nil {
		print(err.Error())
		return nil, err
	}
	print("attached new remote; id=" + newRemote.ID + "; address=" + newRemote.Address)
	return newRemote, nil
}

// GetAllAttached returns all remotes in the remotes map in an array
func (r *Router) GetAllAttached() []*RemoteServer {
	ret := []*RemoteServer{}
	for _, v := range r.remotes {
		ret = append(ret, v)
	}
	return ret
}

// Shutdown marks the remoteServer as shutdown
func (r *Router) Shutdown(id string) {
	print("marking shutdown; id=" + id)
	remote := r.remotes[id]
	if remote == nil {
		return
	}
	r.mu.Lock()
	remote.State = Shutdown
	r.mu.Unlock()
}

// addToMap the remote server, otherwise return error
func (r *Router) addToMap(rm *RemoteServer) error {
	if _, ok := r.remotes[rm.ID]; ok {
		print("could not add to map, id error")
		return errors.New("router.addToMap.error")
	}
	print("added new router to map=" + rm.ID)

	r.mu.Lock()
	r.remotes[rm.ID] = rm
	r.mu.Unlock()

	return nil
}

// removeRemote from the map
func (r *Router) removeRemote(remoteID string) {
	print("removing " + remoteID)
	delete(r.remotes, remoteID)
}

// assignID returns the next ID and then increments the counter
func (r *Router) assignID() string {
	defer func() {
		r.mu.Lock()
		r.idCounter++
		r.mu.Unlock()
	}()
	return "r" + strconv.Itoa(r.idCounter)
}

// State controls the state of a remote server
type State int

const (
	// Running as normal
	Running State = 1 + iota

	// Shutdown notice received from remote
	Shutdown

	// Failed to return a pong
	Failed
)

var states = [...]string{
	"Running",
	"Shutdown",
	"Failed",
}

func (s State) String() string {
	if Running <= s && s <= Failed {
		return states[s-1]
	}
	return "%!State()"
}

// RemoteServer represents the remote process' server
type RemoteServer struct {
	Address string
	ID      string
	State   State

	// Mode determines whether the remote was launched as part of service launcher
	// or it is standalone
	Mode string

	// The time that the state was updated to either Shutdown for Failed
	StateUpdate time.Time

	// The time that the last ping was recorded
	LastPing time.Time

	mu *sync.Mutex

	// fingerprint is assigned and used for ping/pong
	fingerprint string
}

// NewRemoteServer returns an instance of a remote server with values set to the parameters
func NewRemoteServer(id, address, mode string) *RemoteServer {
	return &RemoteServer{
		ID:          id,
		Address:     address,
		State:       Running,
		Mode:        mode,
		LastPing:    time.Now().Add(time.Hour),
		fingerprint: xid.New().String(),
		mu:          &sync.Mutex{},
	}
}

// UpdateState changes the state of the remote server object
func (rs *RemoteServer) UpdateState(newState State) {
	rs.mu.Lock()
	rs.State = newState
	rs.mu.Unlock()
}
