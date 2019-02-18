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
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
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

	fingerprint string
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

	// config data from file
	conf *config.ServiceConfig

	// rpc connection handler to gmbhCore over Cabal
	con *rpc.Connection

	// The user configurable options of the server
	opts options

	// The map that handles function from the user's service
	registeredFunctions map[string]HandlerFunc

	// pingHelper keeps track of channels
	pingHelpers []*pingHelper

	PongTime time.Duration

	// parentID is used only when running inside of a remotepm
	parentID string

	msgCounter int
	mu         *sync.Mutex

	// signalMode chooses between signint and sigusr2 for the shutdown listener
	// depending how how SERVICEMODE environment variable is set
	//
	// sigusr 2 is used only if SERVICEMODE=managed and is intended to only be used
	// in combination with gmbhServiceLauncher
	signalMode string

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
func NewClient(configPath string, opt ...Option) (*Client, error) {

	// Make sure you can't reset the service
	if g != nil {
		return g, nil
	}

	g = &Client{
		registeredFunctions: make(map[string]HandlerFunc),
		mu:                  &sync.Mutex{},
		pingHelpers:         []*pingHelper{},
		PongTime:            time.Second * 45,
		signalMode:          os.Getenv("SERVICEMODE"),
		parentID:            os.Getenv("REMOTE"),
	}

	if g.signalMode == "" {
		g.signalMode = "free"
	}

	g.opts = defaultOptions
	for _, o := range opt {
		o(&g.opts)
	}

	if g.opts.runtime.Verbose {
		notify.SetHeader("[gmbh]")
	}

	// Parse the config either from the path passed in or the one set by the service
	// launcher in the environment
	var err error
	g.conf, err = config.ParseServiceConfig(configPath)
	g.printer("config path=" + configPath)
	if err != nil {
		return nil, errors.New("could not parse the config file; err=" + err.Error())
	}

	// Validate the data from the config file
	err = g.conf.Static.Validate()
	if err != nil {
		return nil, err
	}

	// Set the log path if one is given in the environment
	if os.Getenv("LOGPATH") != "" && os.Getenv("LOGNAME") != "" {
		path := os.Getenv("LOGPATH")
		filename := os.Getenv("LOGNAME") + "-client.log"
		g.outputFile, err = notify.GetLogFileWithPath(path, filename)
		g.outputmu = &sync.Mutex{}
		os.Stdout = g.outputFile
		os.Stderr = g.outputFile
		if err != nil {
			g.printer("could not create log at path=%s", filepath.Join(path+filename))
		}
		g.printer(notify.SEP)
		g.printer("log created")
	} else {
		g.printer("printing all output to stdOut")
	}

	// Set the address to the core if it is given as an environment variable
	// set by the service launcher or otherwise make sure that the default address is used
	if os.Getenv("GMBHCORE") != "" {
		g.conf.CoreAddress = os.Getenv("GMBHCORE")
		g.printer("using core address from env=%s", g.conf.CoreAddress)
	} else if g.conf.CoreAddress == "" {
		g.printer("warning: core_address not set in config, using default=%s", config.DefaultServiceConfig.CoreAddress)
		g.conf.CoreAddress = config.DefaultServiceConfig.CoreAddress
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

	if g.signalMode == "managed" {
		g.printer("managed mode; ignoring siging; listening for sigusr2")
		signal.Ignore(syscall.SIGINT)
		signal.Notify(sigs, syscall.SIGUSR2)
	} else {
		signal.Notify(sigs, syscall.SIGINT)
	}

	g.printer("started, time=" + time.Now().Format(time.RFC3339))

	go g.connect()

	_ = <-sigs
	g.Shutdown(true, "signal")
}

// Shutdown starts shutdown procedures
func (g *Client) Shutdown(forceExit bool, src string) {
	g.printer("Shutdown procedures started in client from " + src)
	g.mu.Lock()
	g.closed = true
	g.reg = nil
	g.pingHelpers = []*pingHelper{}
	g.mu.Unlock()

	g.makeUnregisterRequest()
	g.disconnect()

	// g.printer("shutdown, time=" + time.Now().Format(time.RFC3339))
	if g.signalMode == "managed" {
		g.printer("os.exit in 3s")
		time.Sleep(time.Second * 3)
		os.Exit(0)
	}
	if forceExit {
		g.printer("force os.exit(0)")
		os.Exit(0)
	}
	return
}

/**********************************************************************************
**** Handling connection to gmbhCore
**********************************************************************************/

// connect to gmbhCore
func (g *Client) connect() {
	g.printer("attempting to connect to gmbh-core")

	// when failed or disconnected, the registration is wiped to make sure that
	// legacy data does not get used, thus if g.reg is not nil, then we can assume
	// that a thread has aready requested and received a valid registration
	// and the current thread can be closed
	if g.reg != nil {
		g.printer("cannot (re)connect reg != nil")
		return
	}

	reg, status := makeEphemeralRegistrationRequest(g.conf.Static.Name, true, true, "")
	for status != nil {
		if status.Error() != "registration.gmbhUnavailable" {
			g.printer("gmbh internal error")
			return
		}

		if g.closed || (g.con != nil && g.con.IsConnected()) {
			return
		}
		g.printer("Could not reach gmbh-core, trying again in 5 seconds")
		time.Sleep(time.Second * 5)
		reg, status = makeEphemeralRegistrationRequest(g.conf.Static.Name, true, true, "")

	}

	g.printer("registration details:")
	g.printer("id=" + reg.id + "; address=" + reg.address + "; fingerprint=" + reg.fingerprint)

	if reg.address == "" {
		g.printer("address not received")
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
		g.printer("gmbh connection error=(" + err.Error() + ")")
		return
	}
	g.printer("connected; coreAddress=(" + reg.address + ")")

	go g.sendPing(ph)

}

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
		g.reg = nil
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

		time.Sleep(g.PongTime)
		g.printer("-> ping")

		select {
		case _ = <-ph.pingChan: // case in which this channel is no longer needed
			g.printer("received chan message in sendPing")
			close(ph.pingChan)
			ph.mu.Lock()
			ph.received = true
			ph.mu.Unlock()
			return
		default: // default operation, wait and send a ping

			if !g.con.IsConnected() {
				return
			}

			client, ctx, can, err := rpc.GetCabalRequest(config.DefaultSystemCore.Address, time.Second*30)
			if err != nil {
				g.printer(err.Error())
			}

			if g.reg == nil {
				g.printer("invalid reg for ping")
				return
			}

			ctx = metadata.AppendToOutgoingContext(
				ctx,
				"sender", g.conf.Static.Name,
				"target", "procm",
				"fingerprint", g.reg.fingerprint,
			)

			pong, err := client.Alive(ctx, &intrigue.Ping{
				Time: time.Now().Format(time.Stamp),
			})
			if err != nil {
				g.failed()
				return
			}
			if pong.GetStatus() == "core.verified" {
				can()
				// g.printer("<- pong")
			} else {
				g.printer("<- pong err=" + pong.GetError())
				g.failed()
				return
			}
		}
	}
}

func (g *Client) makeUnregisterRequest() {
	client, ctx, can, err := rpc.GetCabalRequest(g.conf.CoreAddress, time.Second*5)
	if err != nil {
		panic(err)
	}
	defer can()
	request := &intrigue.ServiceUpdate{
		Request: "shutdown.notif",
		Message: g.conf.Static.Name,
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

func handleDataRequest(req intrigue.Request) (*intrigue.Responder, error) {

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
func (r *Request) toProto() *intrigue.Request {
	return &intrigue.Request{
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
func (r *Responder) toProto() *intrigue.Responder {
	return &intrigue.Responder{
		Result:      r.Result,
		ErrorString: r.ErrorString,
		HadError:    r.HadError,
	}
}

// requestFromProto takes a gproto request and returns the corresponding
// Request object
func requestFromProto(r intrigue.Request) Request {
	return Request{
		Sender: r.Sender,
		Target: r.Target,
		Method: r.Method,
		Data1:  r.Data1,
	}
}

// ResponderFromProto takes a gproto Responder and returns the corresponding
// Responder object
func responderFromProto(r intrigue.Responder) Responder {
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
	g.printer("removed " + strconv.Itoa(len(phs)-c) + "/" + strconv.Itoa(len(phs)) + " channels")
	return n
}

func (g *Client) printer(msg string, a ...interface{}) {
	if g.outputFile != nil {
		g.outputmu.Lock()
		g.outputFile.WriteString(fmt.Sprintf(msg, a...) + "\n")
		g.outputmu.Unlock()
	} else {
		if g.opts.runtime.Verbose {
			notify.LnCyanF(msg, a...)
		}
	}
}

/**********************************************************************************
** RPC Functions
**********************************************************************************/

/**********************************************************************************
** RPCClient
**********************************************************************************/

func getRPCClient() (intrigue.CabalClient, error) {
	con, err := grpc.Dial(g.conf.CoreAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return intrigue.NewCabalClient(con), nil

}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest() (intrigue.CabalClient, context.Context, context.CancelFunc, error) {
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

	request := intrigue.NewServiceRequest{
		Service: &intrigue.NewService{
			Name:     name,
			Aliases:  []string{},
			IsClient: isClient,
			IsServer: isServer,
		},
	}

	reply, err := client.RegisterService(ctx, &request)
	if err != nil {
		if grpc.Code(err) == codes.Unavailable {
			return nil, errors.New("registration.gmbhUnavailable")
		}
		g.printer(grpc.Code(err).String())
		return nil, errors.New("registration.gmbhUnavailable")
	}

	if reply.Message == "acknowledged" {

		reg := reply.GetServiceInfo()

		r := &registration{
			id:          reg.GetID(),
			address:     reg.GetAddress(),
			fingerprint: reg.GetFingerprint(),
		}
		return r, nil
	}
	return nil, errors.New(reply.GetMessage())
}

func makeDataRequest(target string, method string, data string) (Responder, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := intrigue.DataRequest{
		Request: &intrigue.Request{
			Sender: g.conf.Static.Name,
			Target: target,
			Method: method,
			Data1:  data,
		},
	}

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	g.printer("<==" + mcs + "== target: " + target + ", method: " + method)

	reply, err := client.Data(ctx, &request)
	if err != nil {
		r := Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return r, err

	}
	g.printer(" ==" + mcs + "==> " + reply.String())
	if reply.Responder == nil {
		return responderFromProto(intrigue.Responder{}), nil
	}
	return responderFromProto(*reply.Responder), nil
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
	intrigue.RegisterCabalServer(s, &_server{})

	reflection.Register(s)
	if err := s.Serve(list); err != nil {
		panic(err)
	}

}

func (s *_server) RegisterService(ctx context.Context, in *intrigue.NewServiceRequest) (*intrigue.Receipt, error) {
	return &intrigue.Receipt{Message: "operation.invalid"}, nil
}

func (s *_server) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {

	g.printer(fmt.Sprintf("-> Update Registration; Message=%s", in.String()))

	request := in.GetRequest()
	// target := in.GetMessage()

	if request == "core.shutdown" {
		g.printer("recieved shutdown")

		g.printer("sending message over chans to ping")
		for _, c := range g.pingHelpers {
			c.pingChan <- true
			c.contacted = true
		}

		g.pingHelpers = update(g.pingHelpers)

		// either shutdown for real or disconnect and try and reach again if
		// the service wasn't forked from gmbh-core
		if g.signalMode == "managed" {
			go g.Shutdown(true, "core")
		} else if !g.closed {
			go func() {

				g.mu.Lock()
				g.reg = nil
				g.mu.Unlock()

				g.disconnect()
				g.connect()
			}()
		}
	}
	return &intrigue.Receipt{Error: "unknown.request"}, nil
}

func (s *_server) Data(ctx context.Context, in *intrigue.DataRequest) (*intrigue.DataResponse, error) {

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	g.printer("==" + mcs + "==> from: " + in.GetRequest().GetSender() + "; method: " + in.GetRequest().GetMethod())

	responder, err := handleDataRequest(*in.GetRequest())
	if err != nil {
		panic(err)
	}
	return &intrigue.DataResponse{Responder: responder}, nil
}

func (s *_server) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {

	g.printer(fmt.Sprintf("-> Summary Request; Action=%s", in.String()))

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		g.printer("Could not get metadata from summary request")
		return &intrigue.SummaryReceipt{Error: "unknown.id"}, nil
	}

	fp := strings.Join(md.Get("fingerprint"), "")
	if fp != g.getReg().fingerprint {
		g.printer("Could not match fingerprint from summary request; incoming fp=%s", fp)
		return &intrigue.SummaryReceipt{Error: "unknown.id"}, nil
	}

	response := &intrigue.SummaryReceipt{
		Services: []*intrigue.CoreService{
			&intrigue.CoreService{
				Name:     g.conf.Static.Name,
				Address:  g.getReg().address,
				Mode:     g.signalMode,
				ParentID: g.parentID,
				Errors:   []string{},
			},
		},
	}

	return response, nil
}

func (s *_server) Alive(ctx context.Context, ping *intrigue.Ping) (*intrigue.Pong, error) {
	return &intrigue.Pong{Time: time.Now().Format(time.Stamp)}, nil
}
