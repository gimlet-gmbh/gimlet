package rpc

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/process"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Connection holds data related to a grpc connection
type Connection struct {
	Server    *grpc.Server
	ctype     string
	Address   string
	Cabal     cabal.CabalServer
	Control   cabal.ControlServer
	Remote    cabal.RemoteServer
	Connected bool
	mu        *sync.Mutex
	Errors    []error
}

// NewCabalConnection returns a new connection object
func NewCabalConnection() *Connection {
	con := newConnection()
	con.ctype = "cabal"
	return &con
}

// NewControlConnection returns a new connection object
func NewControlConnection() Connection {
	con := newConnection()
	con.ctype = "control"
	return con
}

// NewRemoteConnection returns a new connection object
func NewRemoteConnection() Connection {
	con := newConnection()
	con.ctype = "remote"
	return con
}

func newConnection() Connection {
	return Connection{
		Connected: false,
		mu:        &sync.Mutex{},
		Errors:    make([]error, 0),
	}
}

// Connect to grpc server
func (c *Connection) Connect() error {

	if c.Address == "" {
		return errors.New("connection.connect.noAddress")
	}

	list, err := net.Listen("tcp", c.Address)
	if err != nil {
		return errors.New("connection.connect.listener=(" + err.Error() + ")")
	}

	go func() {
		c.Server = grpc.NewServer()

		if c.ctype == "cabal" {
			cabal.RegisterCabalServer(c.Server, c.Cabal)
		} else if c.ctype == "control" {
			cabal.RegisterControlServer(c.Server, c.Control)
		} else if c.ctype == "remote" {
			cabal.RegisterRemoteServer(c.Server, c.Remote)
		}

		reflection.Register(c.Server)

		if err := c.Server.Serve(list); err != nil {
			c.Errors = append(c.Errors, err)
			c.Connected = false
		}

	}()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Connected = true
	return nil
}

// Disconnect from grpc server
func (c *Connection) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Server != nil {
		c.Server.Stop()
	}
	c.Connected = false
}

// GetCabalRequest returns a cabal client to make requests through at address and with timeout
func GetCabalRequest(address string, timeout time.Duration) (cabal.CabalClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, can := context.WithTimeout(context.Background(), timeout)
	return cabal.NewCabalClient(con), ctx, can, nil
}

// GetControlRequest returns a control client to make requests through at address and with timeout
func GetControlRequest(address string, timeout time.Duration) (cabal.ControlClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, can := context.WithTimeout(context.Background(), timeout)
	return cabal.NewControlClient(con), ctx, can, nil
}

// GetRemoteRequest returns a remote client to make requests through at address and with timeout
func GetRemoteRequest(address string, timeout time.Duration) (cabal.RemoteClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, can := context.WithTimeout(context.Background(), timeout)
	return cabal.NewRemoteClient(con), ctx, can, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
// Converters
//////////////////////////////////////////////////////////////////////////////////////////

// ServicesToRPCs translates an array of service pointers to an array of cabal service pointers
func ServicesToRPCs(ss []*service.Service) []*cabal.Service {
	ret := []*cabal.Service{}
	for _, s := range ss {
		ret = append(ret, ServiceToRPC(*s))
	}
	return ret
}

// ServiceToRPC translates one service to cabal form
func ServiceToRPC(s service.Service) *cabal.Service {

	procRuntime := s.GetProcess().GetRuntime()

	rpcService := &cabal.Service{
		Id:      s.ID,
		Name:    s.Static.Name,
		Path:    s.Path,
		LogPath: s.Path + defaults.SERVICE_LOG_PATH + defaults.SERVICE_LOG_FILE,
	}

	if s.Mode == service.Managed {

		rpcService.Pid = int32(procRuntime.Pid)
		rpcService.Fails = int32(procRuntime.Fails)
		rpcService.Restarts = int32(procRuntime.Restarts)
		rpcService.StartTime = procRuntime.StartTime.Format(time.RFC3339)
		rpcService.FailTime = procRuntime.DeathTime.Format(time.RFC3339)
		rpcService.Errors = s.GetProcess().ReportErrors()

		rpcService.Mode = "managed"
		switch s.Process.GetStatus() {
		case process.Stable:
			rpcService.Status = "Stable"
		case process.Running:
			rpcService.Status = "Running"
		case process.Degraded:
			rpcService.Status = "Degraded"
		case process.Failed:
			rpcService.Status = "Failed"
		case process.Killed:
			rpcService.Status = "Killed"
		case process.Initialized:
			rpcService.Status = "Initialized"
		}
	} else if s.Mode == service.Remote {
		rpcService.Mode = "remote"
		rpcService.Status = "-"
	}
	return rpcService
}

func serviceToStruct() *service.Service {
	return nil
}
