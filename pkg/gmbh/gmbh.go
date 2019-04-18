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
	"strings"
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

	PongTime time.Duration

	// the address of the cabal server that the client hosts itself on.
	// This member var is mostly used for use in "C" env mode, or containerized
	myAddress string

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

	// how to handle signals as set by the environment
	// {M,C,""}
	// M = managed; use sigusr
	// C = containerized
	// "" = standalone
	env string

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
		PongTime:            time.Second * 45,
		env:                 os.Getenv("ENV"),
		parentID:            os.Getenv("REMOTE"),
	}

	g.opts = defaultOptions
	for _, o := range opt {
		o(&g.opts)
	}

	if g.opts.service.Name == "" {
		return nil, fmt.Errorf("must set ServiceOptions to include a name for the service")
	}

	print("                    _                 ")
	print("  _  ._ _  |_  |_| /  | o  _  ._ _|_  ")
	print(" (_| | | | |_) | | \\_ | | (/_ | | |_ ")
	print("  _|                                  ")
	print("service started from %s", getpwd())
	print("PeerGroup=" + strings.Join(g.opts.service.PeerGroups, ","))

	// If the address back to core has been set using an environment variable, use that. Otherwise
	// use the one from opts which defaults to the default set from the config package
	if g.env == "C" {
		g.opts.standalone.CoreAddress = os.Getenv("CORE")
		print("using core address from env=%s", os.Getenv("CORE"))
		g.myAddress = os.Getenv("ADDR")
	} else {
		print("core address=%s", g.opts.standalone.CoreAddress)
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

	if g.env == "M" {
		print("managed mode; ignoring sigint; listening for sigusr2")
		signal.Ignore(syscall.SIGINT)
		signal.Notify(sigs, syscall.SIGQUIT)
	} else {
		signal.Notify(sigs, syscall.SIGINT)
	}

	print("started, time=" + time.Now().Format(time.RFC3339))

	go g.connect()

	_ = <-sigs
	g.Shutdown("signal")
}

// Shutdown starts shutdown procedures
func (g *Client) Shutdown(src string) {
	// print("Shutdown procedures started in client from " + src)
	g.mu.Lock()
	g.closed = true
	g.reg = nil
	g.mu.Unlock()

	g.makeUnregisterRequest()
	g.disconnect()

	// print("shutdown, time=" + time.Now().Format(time.RFC3339))
	print("shutdown complete...")
	defer os.Exit(0)
}

func (g *Client) resolveAddress(target string) string {

	addr, ok := g.whoIs[target]

	// address already stored in whoIs map
	if ok {
		return addr
	}

	// ask the core for the address
	print("getting address for " + target)

	err := makeWhoIsRequest(target)
	if err == nil {
		return g.whoIs[target]
	}

	// send through to see if the core can process the request for us
	return g.opts.standalone.CoreAddress
}

/**********************************************************************************
**** Handling connection to gmbhCore
**********************************************************************************/

// disconnect from gmbh-core and go back into connecting mode
func (g *Client) disconnect() {

	print("disconnecting from gmbh-core")

	g.mu.Lock()
	if g.con != nil {
		print("con exists; can send formal disconnect")
		g.con.Disconnect()
		g.con.Server = nil
		g.con.SetAddress("-")
	} else {
		print("con is nil")
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
	print("failed to receive pong; disconnecting")

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
		print("nil reg err")
		return &registration{}
	}
	return g.reg
}

// logStamp is the date format string for log messages
const logStamp = "06/01/02 15:04"

func print(format string, a ...interface{}) {
	name := "client"
	if g.opts.service.Name != "" {
		name = g.opts.service.Name
	}
	format = "[" + time.Now().Format(logStamp) + "] [" + name + "] " + format
	notify.LnMagentaF(format, a...)
	// fmt.Printf(format, a...)
}

// getpwd returns the directory that the process was launched from according to the os package
// Unlike the os package it never returns and error, only an empty string
func getpwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
