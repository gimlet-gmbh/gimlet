package main

import (
	"context"
	"errors"
	"strconv"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/rpc"
)

/**
 * control.go
 * Implements the gRPC server and limited client for the gmbhCore Control Server
 * Abe Dick
 * January 2019
 */

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

type controlServer struct{}

func (c *controlServer) StartService(ctx context.Context, in *cabal.StartRequest) (*cabal.StartReply, error) {
	// TODO: Implement
	return nil, nil
}

func (c *controlServer) RestartService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}

	target, err := cc.Router.LookupByID(in.GetId())
	if err != nil {
		return &cabal.StatusReply{Status: "could not find service: " + err.Error()}, nil
	}
	pid, err := target.Process.Restart(false)
	if err != nil {
		return &cabal.StatusReply{Status: "could not restart service: " + err.Error()}, nil
	}
	return &cabal.StatusReply{Status: "pid=" + strconv.Itoa(pid)}, nil
}

func (c *controlServer) KillService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	// TODO: Implement
	return nil, nil
}

func (c *controlServer) ListAll(ctx context.Context, in *cabal.AllRequest) (*cabal.ListReply, error) {

	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}

	serviceNames := cc.Router.Names
	reply := cabal.ListReply{
		Length:   int32(len(serviceNames)),
		Services: rpc.ServicesToRPCs(cc.Router.GetAllServices()),
	}

	return &reply, nil
}
func (c *controlServer) ListOne(ctx context.Context, in *cabal.SearchRequest) (*cabal.ListReply, error) {

	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}

	target, err := cc.Router.LookupByID(in.GetId())
	if err != nil {
		return &cabal.ListReply{Length: 0}, nil
	}

	reply := cabal.ListReply{
		Length:   1,
		Services: []*cabal.Service{rpc.ServiceToRPC(*target)},
	}

	return &reply, nil
}

func (c *controlServer) RestartAll(ctx context.Context, in *cabal.AllRequest) (*cabal.StatusReply, error) {
	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}
	go cc.Router.RestartAllServices()
	return &cabal.StatusReply{Status: "success"}, nil
}

func (c *controlServer) KillAll(ctx context.Context, in *cabal.AllRequest) (*cabal.StatusReply, error) {
	// TODO: Implement
	return nil, nil
}

func (c *controlServer) StopServer(ctx context.Context, in *cabal.StopRequest) (*cabal.StatusReply, error) {
	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}
	go cc.shutdown(true)
	return &cabal.StatusReply{Status: "shutdown procedure started"}, nil
}

func (c *controlServer) ServerStatus(ctx context.Context, in *cabal.StatusRequest) (*cabal.StatusReply, error) {
	return &cabal.StatusReply{Status: "awk"}, nil
}

/////////////////////////////////////////////////////////////////////////
// CLIENT
/////////////////////////////////////////////////////////////////////////
