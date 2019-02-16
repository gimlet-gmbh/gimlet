package rpc

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/rpc/intrigue"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Connection holds data related to a grpc connection
type Connection struct {
	Server    *grpc.Server
	ctype     string
	Address   string
	Cabal     cabal.CabalServer
	Control   intrigue.ControlServer
	Remote    intrigue.RemoteServer
	Connected bool
	mu        *sync.Mutex
	Errors    []error
}

// NewCabalConnection returns a new connection object
func NewCabalConnection(addr string, server cabal.CabalServer) *Connection {
	con := newConnection()
	con.Address = addr
	con.Cabal = server
	con.ctype = "cabal"
	return &con
}

// NewControlConnection returns a new connection object
func NewControlConnection(addr string, server intrigue.ControlServer) *Connection {
	con := newConnection()
	con.Address = addr
	con.Control = server
	con.ctype = "control"
	return &con
}

// NewRemoteConnection returns a new connection object
func NewRemoteConnection(addr string, server intrigue.RemoteServer) *Connection {
	con := newConnection()
	con.Address = addr
	con.Remote = server
	con.ctype = "remote"
	return &con
}

func newConnection() Connection {
	return Connection{
		Connected: false,
		mu:        &sync.Mutex{},
		Errors:    make([]error, 0),
	}
}

// SetAddress to addr
func (c *Connection) SetAddress(addr string) {
	c.Address = addr
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

		// parms := keepalive.ServerParameters{
		// 	Time:    time.Second * 30,
		// 	Timeout: time.Second * 15,
		// }
		// a := grpc.KeepaliveParams(parms)
		c.Server = grpc.NewServer()

		if c.ctype == "cabal" {
			cabal.RegisterCabalServer(c.Server, c.Cabal)
		} else if c.ctype == "control" {
			intrigue.RegisterControlServer(c.Server, c.Control)
		} else if c.ctype == "remote" {
			intrigue.RegisterRemoteServer(c.Server, c.Remote)
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

// IsConnected to grpc server
func (c *Connection) IsConnected() bool {
	return c.Connected
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
func GetControlRequest(address string, timeout time.Duration) (intrigue.ControlClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, can := context.WithTimeout(context.Background(), timeout)
	return intrigue.NewControlClient(con), ctx, can, nil
}

// GetRemoteRequest returns a remote client to make requests through at address and with timeout
func GetRemoteRequest(address string, timeout time.Duration) (intrigue.RemoteClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, can := context.WithTimeout(context.Background(), timeout)
	return intrigue.NewRemoteClient(con), ctx, can, nil
}

// RemotesToCabal converter
func RemotesToCabal(remotes []*intrigue.ProcessManager) []*cabal.ProcessManager {
	ret := []*cabal.ProcessManager{}
	for _, r := range remotes {
		pm := &cabal.ProcessManager{
			ID:       r.ID,
			Name:     r.Name,
			Address:  r.Address,
			Services: RemoteServicesToCabal(r.Services),
		}
		ret = append(ret, pm)
	}
	return ret
}

// RemoteServicesToCabal converter
func RemoteServicesToCabal(services []*intrigue.Service) []*cabal.Service {
	ret := []*cabal.Service{}
	for _, s := range services {
		ns := &cabal.Service{
			Id:         s.Id,
			Name:       s.Name,
			Registered: s.Registered,
			Mode:       s.Mode,
			Address:    s.Address,
			Path:       s.Path,
			LogPath:    s.LogPath,
			Status:     s.Status,
			Restarts:   s.Restarts,
			Fails:      s.Fails,
			Pid:        s.Pid,
			StartTime:  s.StartTime,
			FailTime:   s.FailTime,
			Errors:     s.Errors,
		}
		ret = append(ret, ns)
	}
	return ret
}
