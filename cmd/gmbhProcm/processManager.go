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

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
)

// Mode controls whether new processes can be attached during runtime or if they must be
// specified before hand in a manifest
type Mode int

const (
	// Dev mode allows processes to be attached during runtime
	Dev Mode = 1 + iota

	// Deploy mode does not allow new processes to be attached during runtime
	Deploy

	// Open mode works as a true process manager and makes no assumptions about the
	// processes it will host
	Open
)

var modes = [...]string{
	"Dev",
	"Deploy",
	"Open",
}

func (m Mode) String() string {
	if Dev <= m && m <= Open {
		return modes[m-1]
	}
	return "%!Mode()"
}

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

	// The mode controls how processes are attached
	mode Mode

	// The connection that hosts the control server
	con *rpc.Connection

	// Router manages all addresses and instances of remotes
	router *Router

	Log     *notify.Log
	ErrLog  *notify.Log
	mu      *sync.Mutex
	verbose bool
}

var procm *ProcessManager

// NewProcessManager instantiates a new pm if one has not already been created. Note that this
// should be assigned to a global instance to interface with the rpc server. The rpc server should
// then use the GetProcM function to ensure that the global has not fallen out of scope.
func NewProcessManager(configFile string, v bool) *ProcessManager {

	// Make sure that it is never allowed to overrite once already instantiated
	if procm != nil {
		return procm
	}

	// TODO: Need to parse the config, for now using defaults

	procm = &ProcessManager{
		Version:   defaults.VERSION,
		CodeName:  defaults.CODE,
		startTime: time.Now(),
		Address:   defaults.PM_ADDRESS,
		router:    NewRouter(),
		mode:      Dev,
		verbose:   v,
		mu:        &sync.Mutex{},
		// Log: notify.NewLogFile()
	}

	notify.LnCyanF("                    _                 ")
	notify.LnCyanF("  _  ._ _  |_  |_| |_) ._ _   _ |\\/| ")
	notify.LnCyanF(" (_| | | | |_) | | |   | (_) (_ |  |  ")
	notify.LnCyanF("  _|                                  ")

	notify.SetHeader("[procm]")

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
	notify.StdMsgBlue("serving at " + p.Address)
	return nil
}

// Wait sets the process manager to block the main thread until shutdown signal is recieved
// either by the terminal or using the control tool
func (p *ProcessManager) Wait() {

	go p.gracefulShutdownListener()

	// set up the listener for shutdown
	sig := make(chan os.Signal, 1)
	if os.Getenv("PMMODE") == "PMManaged" {
		notify.StdMsgBlue("overriding sigusr2")
		notify.StdMsgBlue("ignoring sigint")
		signal.Notify(sig, syscall.SIGUSR2)
		signal.Ignore(syscall.SIGINT)
	} else {
		notify.StdMsgBlue("overriding sigint")
		signal.Notify(sig, syscall.SIGINT)
	}

	notify.StdMsgBlue("main thread waiting...")

	_ = <-sig
	fmt.Println() // deadline to align output after sigint

	p.Shutdown(false)
}

// gracefulShutdownListener
func (p *ProcessManager) gracefulShutdownListener() {

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGUSR1)

	_ = <-shutdown
	notify.StdMsgBlue("SIGUSR1 reported")
	p.sendGmbhShutdown()
}

// RegisterRemote adds the remote to the router and sends back the id and address
func (p *ProcessManager) RegisterRemote() (id, address string, err error) {
	notify.StdMsgBlue("registering new remote")
	rm, err := p.router.AttachNewRemote()
	if err != nil {
		return "", "", err
	}
	return rm.ID, rm.Address, nil
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
func (p *ProcessManager) RestartRemote(id string) error {
	r, err := p.router.LookupRemote(id)
	if err != nil {
		return nil
	}
	return p.sendRestart(r)
}

// RestartAll sends an rpc restart request to all remotes
func (p *ProcessManager) RestartAll() []error {
	all := p.router.GetAllAttached()
	errors := []error{}
	for _, r := range all {
		err := p.sendRestart(r)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		for _, e := range errors {
			notify.StdMsgErr("restart error=" + e.Error())
		}
	} else {
		notify.StdMsgBlue("sent all restart requests with no errors")
	}
	return errors
}

// sendRestart sends a restart request to a remote
func (p *ProcessManager) sendRestart(remote *RemoteServer) error {
	notify.StdMsgBlue("sending restart request to " + remote.ID)

	client, ctx, can, err := rpc.GetRemoteRequest(remote.Address, time.Second*2)
	if err != nil {
		return err
	}
	action := &cabal.Action{
		Sender:  "gmbh-core",
		Target:  remote.ID,
		Action:  "service.restart",
		Message: "all",
	}
	_, err = client.RequestRemoteAction(ctx, action)
	if err != nil {
		return err
	}
	can()
	return nil
}

// sendGmbhShutdown sends a message to all remotes to no longer restart their attached
// processes because gmbh process manager has signaled shutdown time
func (p *ProcessManager) sendGmbhShutdown() {

	notify.LnBlueF("gmbh shutdown initiated")

	remotes := p.router.GetAllAttached()
	for _, r := range remotes {
		notify.StdMsgBlue("sending gmbh shutdown notice to " + r.ID)
		client, ctx, can, err := rpc.GetRemoteRequest(r.Address, time.Second*2)
		if err != nil {
			return
		}
		update := &cabal.ServiceUpdate{
			Sender: "gmbh-core",
			Target: r.ID,
			Action: "gmbh.shutdown",
		}
		_, err = client.UpdateServiceRegistration(ctx, update)
		if err != nil {
			return
		}
		can()
	}
}

// sendShutdown notice to all attached remotes
func (p *ProcessManager) sendShutdown() {
	remotes := p.router.GetAllAttached()
	for _, r := range remotes {
		notify.StdMsgBlue("sending shutdown notice to " + r.ID)
		client, ctx, can, err := rpc.GetRemoteRequest(r.Address, time.Second*2)
		if err != nil {
			return
		}
		update := &cabal.ServiceUpdate{
			Sender: "gmbh-core",
			Target: r.ID,
			Action: "core.shutdown",
		}
		_, err = client.UpdateServiceRegistration(ctx, update)
		if err != nil {
			return
		}
		can()
	}
}

// MarkShutdown marks the remote as having shutdown and being inactive
func (p *ProcessManager) MarkShutdown(id string) {
	p.router.Shutdown(id)
}

// Shutdown starts shutdown procedures. If remote it indicates tat the signal came from the control
// tool
func (p *ProcessManager) Shutdown(remote bool) {
	notify.StdMsgBlue("shutdown signal received")
	p.con.Disconnect()
	p.sendShutdown()
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

// Router controls the handling of attached remote servers including assigning addresses
// and managing the actual instance of each along with reporting all associated errors
type Router struct {
	// map[ID]RemoteServer holds the internal lookup of each of the attached remote servers
	remotes map[string]*RemoteServer

	// idCounter is the next id to assign
	idCounter int

	// handles the assignment of addresses for new remotes
	addressing *addressHandler

	mu *sync.Mutex

	Verbose bool
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	r := &Router{
		remotes:   make(map[string]*RemoteServer),
		idCounter: 100,
		addressing: &addressHandler{
			host: defaults.LOCALHOST,
			port: defaults.PM_START,
			mu:   &sync.Mutex{},
		},
		mu:      &sync.Mutex{},
		Verbose: true,
	}
	// start the ping handler
	go r.pingHandler()

	return r
}

// LookupRemote scans through the remote map and returns if a match is found, otherwise an
// an error is returned
func (r *Router) LookupRemote(id string) (*RemoteServer, error) {
	// r.verbose("looking up remote with id=" + id)
	rm := r.remotes[id]
	if rm == nil {
		r.verbose("attempted to lookkup up remote with id=" + id + "; not found")
		return nil, errors.New("router.LookupRemote.notFound")
	}
	// r.verbose("found remote")
	return rm, nil
}

// AttachNewRemote adds the remote to the map and then returns the new remote server object
func (r *Router) AttachNewRemote() (*RemoteServer, error) {
	newRemote := NewRemoteServer(r.assignID(), r.addressing.Assign())
	err := r.addToMap(newRemote)
	if err != nil {
		r.verbose(err.Error())
		return nil, err
	}
	r.verbose("attached new remote; id=" + newRemote.ID + "; address=" + newRemote.Address)
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
	r.verbose("marking shutdown; id=" + id)
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
		r.verbose("could not add to map, id error")
		return errors.New("router.addToMap.error")
	}
	r.verbose("added new router to map=" + rm.ID)

	r.mu.Lock()
	r.remotes[rm.ID] = rm
	r.mu.Unlock()

	return nil
}

// pingHandler looks through each of the remotes in the map. if it has been more than n amount of
// time since a remote has sent a ping, it will be pinged. If the ping is not retured after n more
// seconds, the remote will be marked as Failed After n amount of time, failed remotes will
// be removed from the map
func (r *Router) pingHandler() {
	for {
		time.Sleep(time.Second * 45)
		// notify.StdMsgBlue("checking for pings")
		for _, v := range r.GetAllAttached() {
			if v.State == Failed {
				if time.Since(v.StateUpdate) > time.Second*30 {
					r.removeRemote(v.ID)
				}

			} else if v.State == Shutdown {
				if time.Since(v.StateUpdate) > time.Second*90 {
					r.removeRemote(v.ID)
				}
			} else if v.State == Running {
				if time.Since(v.LastPing) > time.Second*90 {
					v.UpdateState(Failed)
				}
			}
		}
	}
}

// removeRemote from the map
func (r *Router) removeRemote(remoteID string) {
	r.verbose("removing " + remoteID)
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

// verbose sends message to notify if in verbose mode
func (r *Router) verbose(msg string) {
	if r.Verbose {
		notify.StdMsgBlueNoPrompt("[rtr] " + msg)
	}
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

	// The time that the state was updated to either Shutdown for Failed
	StateUpdate time.Time

	// The time that the last ping was recorded
	LastPing time.Time

	mu *sync.Mutex
}

// NewRemoteServer returns an instance of a remote server with values set to the parameters
func NewRemoteServer(id, address string) *RemoteServer {
	return &RemoteServer{
		ID:       id,
		Address:  address,
		State:    Running,
		LastPing: time.Now().Add(time.Hour),
		mu:       &sync.Mutex{},
	}
}

// UpdateState changes the state of the remote server object
func (rs *RemoteServer) UpdateState(newState State) {
	rs.mu.Lock()
	rs.State = newState
	rs.mu.Unlock()
}

// UpdatePing marks the time
func (rs *RemoteServer) UpdatePing(t time.Time) {
	rs.mu.Lock()
	rs.LastPing = t
	rs.mu.Unlock()
}

// addressHandler is in charge of assigning addresses to services
type addressHandler struct {
	table map[string]string
	host  string
	port  int
	mu    *sync.Mutex
}

// Assign returns the next address
func (a *addressHandler) Assign() string {
	a.setNextAddress()
	addr := a.host + ":" + strconv.Itoa(a.port)
	return addr
}

// setNextAddress calculates the next address
func (a *addressHandler) setNextAddress() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.port += 2
}
