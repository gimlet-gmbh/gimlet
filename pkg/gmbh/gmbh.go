package gmbh

/*
 * gmbh.go
 * Abe Dick
 * Nov 2018
 */

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	yaml "gopkg.in/yaml.v2"
)

// HandlerFunc is the publically exposed function to register and use the callback functions
// from within gmbhCore. Its behavior is modeled after the http handler that is baked into go
// by default
type HandlerFunc = func(req Request, resp *Responder)

// Client - the structure between a service and CORE
type Client struct {
	conf                *config
	configured          bool
	blocking            bool
	address             string
	registeredFunctions map[string]HandlerFunc
	msgCounter          int
}

// config is the data structure to hold the user config settings for a service
type config struct {
	ServiceName string   `yaml:"name"`
	Aliases     []string `yaml:"aliases"`
	IsServer    bool     `yaml:"isserver"`
	IsClient    bool     `yaml:"isclient"`
	CoreAddress string   `yaml:"core_address"`
}

// g - the gmbhCore object that contains the parsed yaml config and other associated data
var g *Client

// NewService should be called only once. It returns the object in which parameters, and
// handler functions can be attached to gmbh Client.
func NewService() *Client {

	// Make sure you can't reset the service
	if g != nil {
		return g
	}

	notify.SetVerbose(false)

	g := &Client{
		registeredFunctions: make(map[string]HandlerFunc),
		configured:          false,
		blocking:            true,
	}
	return g

}

// Config specifies a config file to use with gmbh client
func (g *Client) Config(path string) (*Client, error) {
	var err error
	g.conf, err = parseYamlConfig(path)
	if err != nil {
		notify.StdMsgErr("could not parse config=" + path)
		return nil, errors.New("could not parse config=" + path)
	}
	g.configured = true
	return g, nil
}

// Verbose runs the client in verbose mode
func (g *Client) Verbose() *Client {
	notify.SetTag("[gmbh-client-debug] ")
	notify.SetVerbose(true)
	return g
}

// Nonblocking runs the client in blocking mode, in otherwords it keeps the service running
// untill a shutdown signal is received. Otherwise the process will exit.
//
// This mode should be used if you are implementing a backend-only service
func (g *Client) Nonblocking() *Client {
	g.blocking = false
	return g
}

// Start registers the service with gmbh in a new goroutine if blocking, else sets the listener and blocks the
// main thread awaiting calls from gRPC.
func (g *Client) Start() {
	if g.blocking {
		g.start()
	} else {
		go g.start()
	}
}

func (g *Client) start() {
	notify.StdMsgNoPrompt("------------------------------------------------------------")
	notify.StdMsg("started, time=" + time.Now().Format(time.RFC3339))
	if g.configured {
		notify.StdMsg("gmbh configuration valid")
	} else {
		notify.StdMsgErr("gmbh configuration invalid")
	}
	addr, err := makeEphemeralRegistrationRequest(g.conf.ServiceName, g.conf.IsClient, g.conf.IsServer)
	if err != nil {
		for err.Error() == "registration.gmbhUnavailable" {
			notify.StdMsgErr("Could not reach gmbhCore, trying again in 5 seconds")
			time.Sleep(time.Second * 5)
			addr, err = makeEphemeralRegistrationRequest(g.conf.ServiceName, g.conf.IsClient, g.conf.IsServer)
		}
		notify.StdMsg("gmbh.Start.error: " + err.Error())
	}
	notify.StdMsgGreen("connected to core=" + g.conf.CoreAddress)
	if addr != "" {
		notify.StdMsgGreen("assigned address=" + addr)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()
	if addr != "" {
		go rpcConnect(addr)
	}

	<-done
	makeUnregisterRequest(g.conf.ServiceName)
	notify.StdMsg("shutdown, time=" + time.Now().Format(time.RFC3339))
	os.Exit(0)
}

// Route - Callback functions to be used when handling data
// requests from gmbh or other services
//
// TODO: Add a mechanism to safely add these and check for collisions, etc.
func (g *Client) Route(route string, handler HandlerFunc) {
	g.registeredFunctions[route] = handler
}

// MakeRequest is the default method for making data requests through gmbh
func (g *Client) MakeRequest(target, method, data string) (Responder, error) {
	resp, err := makeDataRequest(target, method, data)
	if err != nil {
		return Responder{}, errors.New("could not complete request: " + err.Error())
	}
	return resp, nil
}

func handleDataRequest(req cabal.Request) (*cabal.Responder, error) {

	var request Request
	request = requestFromProto(req)
	responder := Responder{}

	handler, ok := g.registeredFunctions[request.Method]
	if !ok {
		responder.HadError = true
		responder.ErrorString = "Could not locate method in registered process map"
	} else {
		handler(request, &responder)
	}

	return responder.toProto(), nil
}

func parseYamlConfig(relativePath string) (*config, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	var conf config
	yamlFile, err := ioutil.ReadFile(path + "/" + relativePath)
	if err != nil {
		notify.StdMsgErr(path + relativePath)
		return nil, errors.New("could not find yaml file")
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, errors.New("could not unmarshal config")
	}
	return &conf, nil
}

// Request is the publically exposed requester between services in gmbh
type Request struct {
	// Sender is the name of the service that is sending the message
	Sender string

	// Target is the name or alias of the intended recepient
	Target string

	// Method is the handler to invoke on target entry
	// TODO: Change this as it can be considered confusing with
	// 		 the HTTP methods...
	Method string

	// Data1 is the data to send
	// TODO: remove this and more articulately handle data
	Data1 string
}

// ToProto returns the gproto Request object corresponding to the current
// Request object
func (r *Request) toProto() *cabal.Request {
	return &cabal.Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}

// Responder is the publically exposed responder between services in gmbh
type Responder struct {
	// Result is the resulting datat from target
	// TODO: remove this and more articulately handle data
	Result string

	// ErrorString is the corresponding error string if HadError is true
	ErrorString string

	// HadError is true if the request was not completed without error
	HadError bool
}

// ToProto returns the gproto Request object corresponding to the current
// Responder object
func (r *Responder) toProto() *cabal.Responder {
	return &cabal.Responder{
		Result:      r.Result,
		ErrorString: r.ErrorString,
		HadError:    r.HadError,
	}
}

// requestFromProto takes a gproto request and returns the corresponding
// Request object
func requestFromProto(r cabal.Request) Request {
	return Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}

// ResponderFromProto takes a gproto Responder and returns the corresponding
// Responder object
func responderFromProto(r cabal.Responder) Responder {
	return Responder{
		Result:      r.Result,
		ErrorString: r.ErrorString,
		HadError:    r.HadError,
	}
}
