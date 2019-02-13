package gmbh

/*
 * gmbh.go
 * Abe Dick
 * Nov 2018
 */

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"

	yaml "gopkg.in/yaml.v2"
)

// HandlerFunc is the publically exposed function to register and use the callback functions
// from within gmbhCore. Its behavior is modeled after the http handler that is baked into go
// by default
type HandlerFunc = func(req Request, resp *Responder)

// Option functions set options from the client
type Option func(*options)

// options contain the runtime configurable parameters
type options struct {

	// RuntimeOptions are options that can be determined at runtime
	runtime *RuntimeOptions
}

// RuntimeOptions - user configurable
type RuntimeOptions struct {
	// Should the client block the main thread until shutdown signal is received?
	Blocking bool

	// Should the client run in verbose mode. in Verbose mode, debug information regarding
	// the gmbh client will be printed to stdOut
	Verbose bool
}

// userconfig are determined in the config file of the service
type userconfig struct {
	// User assigned name
	ServiceName string `yaml:"name"`

	// User assigned aliases
	Aliases []string `yaml:"aliases"`

	// Makes requests to other services
	IsClient bool `yaml:"is_client"`

	// the address back to core
	CoreAddress string `yaml:"core_address"`

	// the intended mode
	Mode string `yaml:"mode"`
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

var defaultOptions = options{
	runtime: &RuntimeOptions{
		Blocking: false,
		Verbose:  false,
	},
}

// SetRuntime options of the client
func SetRuntime(r RuntimeOptions) Option {
	return func(o *options) {
		o.runtime.Blocking = r.Blocking
		o.runtime.Verbose = r.Verbose
	}
}

// Client - the structure between a service and gmbhCore
type Client struct {

	// registratrion with data from gmbhCore
	reg *registration

	// static config data from file
	conf *userconfig

	// rpc connection handler to gmbhCore over Cabal
	con *rpc.Connection

	// The user configurable options of the server
	opts options

	// The map that handles function from the user's service
	registeredFunctions map[string]HandlerFunc

	// pingHelper keeps track of channels
	pingHelpers []*pingHelper

	msgCounter int
	mu         *sync.Mutex

	// closed is set true when shutdown procedures have been started
	closed bool
}

// g - the gmbhCore object that contains the parsed yaml config and other associated data
var g *Client

// NewClient should be called only once. It returns the object in which parameters, and
// handler functions can be attached to gmbh Client.
func NewClient(configPath string, opt ...Option) (*Client, error) {

	// Make sure you can't reset the service
	if g != nil {
		return g, nil
	}

	g = &Client{
		registeredFunctions: make(map[string]HandlerFunc),
		mu:                  &sync.Mutex{},
		pingHelpers:         []*pingHelper{},
	}

	g.opts = defaultOptions
	for _, o := range opt {
		o(&g.opts)
	}

	if g.opts.runtime.Verbose {
		notify.SetVerbose(true)
		notify.SetTag("[cli] ")
	}

	var err error
	g.conf, err = parseConfig(configPath)
	if err != nil {
		return nil, errors.New("could not parse the config file")
	}

	err = validConfig(g.conf)
	if err != nil {
		return nil, err
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
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()

	notify.StdMsgNoPrompt("------------------------------------------------------------")
	notify.StdMsg("started, time=" + time.Now().Format(time.RFC3339))

	go g.connect()

	<-done

	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()

	g.disconnect()

	makeUnregisterRequest(g.conf.ServiceName)
	notify.StdMsg("shutdown, time=" + time.Now().Format(time.RFC3339))
	os.Exit(0)
}

func parseConfig(relativePath string) (*userconfig, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	var stat userconfig
	yamlFile, err := ioutil.ReadFile(path + "/" + relativePath)
	if err != nil {
		notify.StdMsgErr(path + relativePath)
		return nil, errors.New("could not find yaml file")
	}
	err = yaml.Unmarshal(yamlFile, &stat)
	if err != nil {
		return nil, errors.New("could not unmarshal config")
	}
	return &stat, nil
}

func validConfig(c *userconfig) error {
	if c.ServiceName == "" {
		return errors.New("service config must contain name")
	}
	if c.CoreAddress == "" {
		return errors.New("service config must contain address to gmbhCore")
	}
	return nil
}

/**********************************************************************************
**** Handling connection to gmbhCore
**********************************************************************************/

// connect to gmbhCore
func (g *Client) connect() {
	notify.StdMsgBlue("attempting to connect to gmbh-core")

	// when failed or disconnected, the registration is wiped to make sure that
	// legacy data does not get used, thus if g.reg is not nil, then we can assume
	// that a thread has aready requested and received a valid registration
	// and the current thread can be closed
	if g.reg != nil {
		return
	}

	reg, status := makeEphemeralRegistrationRequest(g.conf.ServiceName, g.conf.IsClient, true, "")
	for status != nil {
		if status.Error() != "registration.gmbhUnavailable" {
			notify.StdMsgErr("gmbh internal error")
			return
		}

		if g.closed || (g.con != nil && g.con.IsConnected()) {
			return
		}
		notify.StdMsgErr("Could not reach gmbh-core, trying again in 5 seconds")
		time.Sleep(time.Second * 5)
		reg, status = makeEphemeralRegistrationRequest(g.conf.ServiceName, g.conf.IsClient, true, "")

	}

	notify.StdMsgBlue("registration details:")
	notify.StdMsgBlue("id=" + reg.id + "; address=" + reg.address)
	notify.StdMsgBlue("mode=" + reg.mode + "; corePath=" + reg.corePath)

	if reg.address == "" {
		notify.StdMsgErr("address not received")
		return
	}

	g.mu.Lock()
	g.reg = reg
	g.con = rpc.NewCabalConnection(reg.address, &_server{})

	// add a new channel to communicate to this goroutine

	ph := newPingHelper()
	g.pingHelpers = append(g.pingHelpers, ph)
	g.mu.Unlock()

	err := g.con.Connect()
	if err != nil {
		notify.StdMsgErr("gmbh connection error=(" + err.Error() + ")")
		return
	}
	notify.StdMsgGreen("connected; coreAddress=(" + reg.address + ")")

	go g.sendPing(ph)

}

// disconnect from gmbh-core and go back into connecting mode
func (g *Client) disconnect() {

	notify.StdMsgBlue("disconnecting from gmbh-core")

	g.mu.Lock()
	if g.con != nil {
		notify.StdMsgErr("con good")
		g.con.Disconnect()
		g.con.Server = nil
		g.con.SetAddress("-")
	} else {
		notify.StdMsgErr("con should not be nil in disconnect")
	}
	g.reg = nil
	g.mu.Unlock()

	if !g.closed {
		time.Sleep(time.Second * 5)
		g.connect()
	}
}

func (g *Client) failed() {
	notify.StdMsg("failed to receive pong; disconnecting")

	if g.con.IsConnected() {
		g.con.Disconnect()
	}
	g.con.Server = nil

	if g.reg.mode == "Managed" {
		os.Exit(1)
	}

	if !g.closed {
		time.Sleep(time.Second * 2)
		g.connect()
	}
}

// sendPing is meant to run in its own thread. It will continue to call itself or
// return and changed the state of the connection if there is a failure reaching
// the control server that is ran by gmbhCore
func (g *Client) sendPing(ph *pingHelper) {

	// Loop forever
	for {

		time.Sleep(time.Second * 45)
		notify.StdMsgBlue("-> ping")

		select {
		case _ = <-ph.pingChan: // case in which this channel is no longer needed
			notify.StdMsgBlue("received chan message in sendPing")
			close(ph.pingChan)
			ph.mu.Lock()
			ph.received = true
			ph.mu.Unlock()
			return
		default: // default operation, wait and send a ping

			if !g.con.IsConnected() {
				return
			}

			client, ctx, can, err := rpc.GetCabalRequest("localhost:59999", time.Second*30)
			if err != nil {
				notify.StdMsgErr(err.Error())
			}

			_, err = client.Alive(ctx, &cabal.Ping{Time: time.Now().Format(time.Stamp)})
			if err == nil {
				can()
				notify.StdMsgBlue("<- pong")
			} else {
				g.failed()
				return
			}
		}
	}
}

/**********************************************************************************
**** Handling Data Requests
**********************************************************************************/

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

/**********************************************************************************
** Helpers
**********************************************************************************/

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
** RPC Functions
**********************************************************************************/

/**********************************************************************************
** RPCClient
**********************************************************************************/

func getRPCClient() (cabal.CabalClient, error) {
	con, err := grpc.Dial("localhost:59999", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return cabal.NewCabalClient(con), nil

}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest() (cabal.CabalClient, context.Context, context.CancelFunc, error) {
	client, err := getRPCClient()
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return client, ctx, can, nil
}

func makeEphemeralRegistrationRequest(name string, isClient bool, isServer bool, mode string) (*registration, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := cabal.RegServReq{
		NewServ: &cabal.NewService{
			Name:     name,
			Aliases:  []string{},
			IsClient: isClient,
			IsServer: isServer,
		},
	}

	if mode == "remote" {
		request.NewServ.Mode = cabal.NewService_REMOTE
	} else {
		request.NewServ.Mode = cabal.NewService_PLANETARY
	}

	reply, err := client.EphemeralRegisterService(ctx, &request)
	if err != nil {
		if grpc.Code(err) == codes.Unavailable {
			return nil, errors.New("registration.gmbhUnavailable")
		}
		panic(err)
	}

	if reply.Status == "acknowledged" {
		r := &registration{
			id:       reply.ID,
			mode:     reply.Mode,
			address:  reply.Address,
			corePath: reply.CorePath,
		}
		return r, nil
	}
	return nil, errors.New(reply.GetStatus())
}

func makeDataRequest(target string, method string, data string) (Responder, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := cabal.DataReq{
		Req: &cabal.Request{
			Sender: "test",
			Target: target,
			Method: method,
			Data1:  data,
		},
	}

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	notify.StdMsgNoPrompt("<==" + mcs + "== target: " + target + ", method: " + method)

	reply, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		// panic(err)
		fmt.Println(fmt.Errorf("%v", err.Error()))

		r := Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return r, err

	}
	notify.StdMsgNoPrompt(" ==" + mcs + "==> result: " + reply.Resp.Result + ", errors?: " + reply.Resp.ErrorString)

	return responderFromProto(*reply.Resp), nil
}

func makeUnregisterRequest(name string) {
	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	_, _ = client.UnregisterService(ctx, &cabal.UnregisterReq{Name: name})
}

/**********************************************************************************
** RPC Server
**********************************************************************************/

// _server implements the coms service using gRPC
type _server struct{}

func rpcConnect(address string) {
	list, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	cabal.RegisterCabalServer(s, &_server{})

	reflection.Register(s)
	if err := s.Serve(list); err != nil {
		panic(err)
	}

}

func (s *_server) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {
	return &cabal.RegServRep{Status: "invalid operation"}, nil
}

func (s *_server) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
	return &cabal.UnregisterResp{Awk: false}, nil
}

func (s *_server) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	notify.StdMsgNoPrompt("==" + mcs + "==> from: " + in.Req.Sender + ", method: " + in.Req.Method)

	responder, err := handleDataRequest(*in.Req)
	if err != nil {
		panic(err)
	}

	reply := &cabal.DataResp{Resp: responder}
	return reply, nil
}

func (s *_server) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {

	response := cabal.QueryResponse{
		Awk:     true,
		Status:  true,
		Details: make(map[string]string),
	}

	return &response, nil
}

func (s *_server) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {

	notify.StdMsgBlue(fmt.Sprintf("-> Update Service Request; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetTarget() != g.conf.ServiceName {
		notify.StdMsgErr("invalid target")
		reply := &cabal.ServiceUpdate{
			Action:  "error",
			Message: "invalid service name",
		}
		return reply, nil
	}

	if in.Action == "core.shutdown" {
		notify.StdMsgBlue("recieved shutdown")
		fmt.Println(g.closed)

		notify.StdMsgBlue("sending message over chans to ping")
		fmt.Println("sending " + strconv.Itoa(len(g.pingHelpers)) + " messages")
		for _, c := range g.pingHelpers {
			c.pingChan <- true
			c.contacted = true
		}

		g.pingHelpers = update(g.pingHelpers)

		if !g.closed {
			go func() {
				g.disconnect()
				g.connect()
			}()
		}
	}

	reply := &cabal.ServiceUpdate{Action: "acknowledged"}
	return reply, nil
}

func (s *_server) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
}
