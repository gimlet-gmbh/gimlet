package rpc

import (
	"errors"
	"net"
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
	Connected bool
	Errors    []error
}

// NewCabalConnection returns a new connection object
func NewCabalConnection() *Connection {
	return &Connection{
		Connected: false,
		ctype:     "cabal",
		Errors:    make([]error, 0),
	}
}

// NewControlConnection returns a new connection object
func NewControlConnection() Connection {
	return Connection{
		Connected: false,
		ctype:     "control",
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
		}

		reflection.Register(c.Server)

		if err := c.Server.Serve(list); err != nil {
			c.Errors = append(c.Errors, err)
			c.Connected = false
		}

	}()

	c.Connected = true
	return nil
}

// Disconnect from grpc server
func (c *Connection) Disconnect() {
	if c.Connected {
		c.Server.GracefulStop()
		c.Connected = false
	}
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
