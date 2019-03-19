package gmbh

/*
 * gmbh.go
 * Abe Dick
 * Nov 2018
 */

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
)

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

	// a unique identifier from core to identify the client with core on requests
	fingerprint string
}

type State int

const (
	Connected    State = 1
	Disconnected State = 2
)

// Client - the structure between a service and gmbhCore
type Client struct {

	// registratrion with data from gmbhCore
	reg *registration

	// The user configurable options of the server
	opts options

	// rpc connection handler to gmbhCore over Cabal
	con *rpc.Connection

	// The map that handles function from the user's service
	registeredFunctions map[string]HandlerFunc

	// pingHelper keeps track of channels
	pingHelpers []*pingHelper

	PongTime  time.Duration
	PingCount int

	state State

	// parentID is used only when running inside of a remotepm
	parentID string

	// whoIs map [name]address
	//
	// if the name is not found in the map, a whois request will be sent to gmbhCore
	// where it will be determined if the service can make the connection. The resulting
	// address will be stored in this map
	whoIs map[string]string

	msgCounter int
	mu         *sync.Mutex

	errors   []string
	warnings []string

	// mode chooses between signint and sigusr2 for the shutdown listener
	// depending how how SERVICEMODE environment variable is set
	//
	// sigusr 2 is used only if SERVICEMODE=managed and is intended to only be used
	// in combination with gmbhServiceLauncher
	mode string

	// if a log path can be determined from the environment, it will be stored here and
	// the printer helper will use it instead of stdOut and stdErr
	outputFile *os.File
	outputmu   *sync.Mutex

	// closed is set true when shutdown procedures have been started
	closed bool
}

// g - the gmbhCore object that contains the parsed yaml config and other associated data
var g *Client

// NewClient should be called only once. It returns the object in which parameters, and
// handler functions can be attached to gmbh Client.
func NewClient(opt ...Option) (*Client, error) {

	// Make sure you can't reset the service
	if g != nil {
		return g, nil
	}

	g = &Client{
		registeredFunctions: make(map[string]HandlerFunc),
		whoIs:               make(map[string]string),
		mu:                  &sync.Mutex{},
		pingHelpers:         []*pingHelper{},
		PongTime:            time.Second * 45,
		mode:                os.Getenv("SERVICEMODE"),
		parentID:            os.Getenv("REMOTE"),
	}

	g.opts = defaultOptions
	for _, o := range opt {
		o(&g.opts)
	}

	if g.opts.service.Name == "" {
		return nil, fmt.Errorf("must set ServiceOptions to include a name for the service")
	}

	g.printer("                    _                 ")
	g.printer("  _  ._ _  |_  |_| /  | o  _  ._ _|_  ")
	g.printer(" (_| | | | |_) | | \\_ | | (/_ | | |_ ")
	g.printer("  _|                                  ")
	notify.SetHeader("[gmbh]")
	g.printer("service started from %s", notify.Getpwd())

	// If the address back to core has been set using an environment variable, use that. Otherwise
	// use the one from opts which defaults to the default set from the config package
	if os.Getenv("GMBHCORE") != "" {
		g.opts.standalone.CoreAddress = os.Getenv("GMBHCORE")
		g.printer("using core address from env=%s", os.Getenv("GMBHCORE"))
	} else {
		g.printer("core address=%s", g.opts.standalone.CoreAddress)
	}

	// the mode is determined if it comes from the environment variable initially, otherwise it is set
	// to unmanaged
	if g.mode == "" {
		g.mode = "unmanaged"
	}

	// @important -- the only service allowed to be named CoreData is the actual gmbhCore
	if g.opts.service.Name == "CoreData" {
		return nil, errors.New("\"CoreData\" is a reserved service name")
	}

	return g, nil
}

/**********************************************************************************
**** Handling Client Operation
**********************************************************************************/

// Start registers the service with gmbh in a new goroutine if blocking, else sets the listener and blocks the
// main thread awaiting calls from gRPC.
func (g *Client) Start() {
	if g.opts.runtime.Blocking {
		g.start()
	} else {
		go g.start()
	}
}

func (g *Client) start() {
	sigs := make(chan os.Signal, 1)

	if g.mode == "managed" {
		g.printer("managed mode; ignoring siging; listening for sigusr2")
		signal.Ignore(syscall.SIGINT)
		signal.Notify(sigs, syscall.SIGQUIT)
	} else {
		signal.Notify(sigs, syscall.SIGINT)
	}

	g.printer("started, time=" + time.Now().Format(time.RFC3339))

	go g.connect()

	_ = <-sigs
	g.Shutdown("signal")
}

// Shutdown starts shutdown procedures
func (g *Client) Shutdown(src string) {
	g.printer("Shutdown procedures started in client from " + src)
	g.mu.Lock()
	g.closed = true
	g.reg = nil
	g.pingHelpers = []*pingHelper{}
	g.mu.Unlock()

	g.makeUnregisterRequest()
	g.disconnect()

	if g.mode == "managed" {
		g.printer("managed shutdown on return")
		defer os.Exit(0)

	}
	g.printer("shutdown, time=" + time.Now().Format(time.RFC3339))
	defer os.Exit(0)
}

/**********************************************************************************
**** Handling connection to gmbhCore
**********************************************************************************/

// disconnect from gmbh-core and go back into connecting mode
func (g *Client) disconnect() {

	g.printer("disconnecting from gmbh-core")

	g.mu.Lock()
	if g.con != nil {
		g.printer("con exists; can send formal disconnect")
		g.con.Disconnect()
		g.con.Server = nil
		g.con.SetAddress("-")
	} else {
		g.printer("con is nil")
	}
	g.reg = nil
	g.state = Disconnected
	g.mu.Unlock()

	if !g.closed {
		time.Sleep(time.Second * 5)
		g.connect()
	}
}

func (g *Client) failed() {
	g.printer("failed to receive pong; disconnecting")

	if g.con.IsConnected() {
		g.con.Disconnect()
	}
	g.con.Server = nil

	if g.reg.mode == "Managed" {
		os.Exit(1)
	}

	if !g.closed {
		g.mu.Lock()
		g.reg = nil
		g.state = Disconnected
		g.mu.Unlock()
		time.Sleep(time.Second * 2)
		g.connect()
	}
}

func (g *Client) makeUnregisterRequest() {
	client, ctx, can, err := rpc.GetCabalRequest(g.opts.standalone.CoreAddress, time.Second*5)
	if err != nil {
		panic(err)
	}
	defer can()
	request := &intrigue.ServiceUpdate{
		Request: "shutdown.notif",
		Message: g.opts.service.Name,
	}
	_, _ = client.UpdateRegistration(ctx, request)
}

// getReg gets the registration or an empty one, keeps from causing a panic
func (g *Client) getReg() *registration {
	if g.reg == nil {
		g.printer("nil reg err")
		return &registration{}
	}
	return g.reg
}

func (g *Client) printer(msg string, a ...interface{}) {
	if g.opts.runtime.Verbose {
		notify.LnMagentaF(msg, a...)
	}
}
