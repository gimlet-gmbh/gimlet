package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
)

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

type controlServer struct{}

func (c *controlServer) StartService(ctx context.Context, in *cabal.StartRequest) (*cabal.StartReply, error) {
	// TODO: Implement
	return &cabal.StartReply{Status: "invalid"}, nil
}

func (c *controlServer) RestartService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	rpcMessage("<- restart one request")

	// make sure that sender is gmbh-ctrl
	if in.GetSender() != "gmbh-ctrl" {
		rpcMessage("reporting invalid sender")
		return &cabal.StatusReply{Status: "invalid sender"}, nil
	}

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.StatusReply{Status: "internal server error"}, nil
	}

	remote, err := pm.LookupRemote(in.GetParentID())
	if err != nil {
		notify.StdMsgErr("could not contact " + in.GetParentID())
		return &cabal.StatusReply{Status: "could not contact remote"}, nil
	}

	pid := "-1"
	{
		client, ctx, can, err := rpc.GetRemoteRequest(remote.Address, time.Second*15)
		if err != nil {
			notify.StdMsgErr("could not contact " + remote.ID)
		}
		request := &cabal.Action{
			Sender:  "gmbh-core",
			Target:  remote.ID,
			Message: in.GetId(),
			Action:  "service.restart",
		}
		reply, err := client.RequestRemoteAction(ctx, request)
		if err != nil {
			notify.StdMsgErr("could not contact " + remote.ID)
		}
		pid = reply.GetStatus()
		can()
	}

	response := &cabal.StatusReply{Status: "pid=" + pid}
	return response, nil

}

func (c *controlServer) KillService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	return nil, nil
}

func (c *controlServer) ListAll(ctx context.Context, in *cabal.AllRequest) (*cabal.ListReply, error) {

	rpcMessage("<- list all request")

	// make sure that sender is gmbh-ctrl
	if in.GetSender() != "gmbh-ctrl" {
		rpcMessage("reporting invalid sender")
		return &cabal.ListReply{Status: "invalid sender"}, nil
	}

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.ListReply{Status: "internal server error"}, nil
	}

	rmts := pm.GetAllRemotes()
	rpcrmts := []*cabal.ProcessManager{}
	for _, re := range rmts {
		rpcServices := []*cabal.Service{}
		{
			client, ctx, can, err := rpc.GetRemoteRequest(re.Address, time.Second*2)
			if err != nil {
				notify.StdMsgErr("could not contact " + re.ID)
				continue
			}
			request := &cabal.Action{
				Sender: "gmbh-core",
				Target: re.ID,
				Action: "request.info.all",
			}
			reply, err := client.RequestRemoteAction(ctx, request)
			if err != nil {
				notify.StdMsgErr("could not contact " + re.ID)
				continue
			}
			rpcServices = append(rpcServices, reply.GetServices()...)
			can()
		}

		rpcrmts = append(rpcrmts, remoteToRPC(re, rpcServices))
	}

	response := &cabal.ListReply{
		Status:  "ack",
		Remotes: rpcrmts,
	}
	return response, nil

}
func (c *controlServer) ListOne(ctx context.Context, in *cabal.SearchRequest) (*cabal.ListReply, error) {

	rpcMessage("<- list one request")

	// make sure that sender is gmbh-ctrl
	if in.GetSender() != "gmbh-ctrl" {
		rpcMessage("reporting invalid sender")
		return &cabal.ListReply{Status: "invalid sender"}, nil
	}

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.ListReply{Status: "internal server error"}, nil
	}

	rmt, err := pm.LookupRemote(in.GetParentID())
	if err != nil {
		return &cabal.ListReply{Status: "Not Found"}, nil
	}
	rpcServices := []*cabal.Service{}
	{
		client, ctx, can, err := rpc.GetRemoteRequest(rmt.Address, time.Second*5)
		if err != nil {
			notify.StdMsgErr("could not contact " + rmt.ID)
		}
		request := &cabal.Action{
			Sender:  "gmbh-core",
			Target:  rmt.ID,
			Message: in.GetId(),
			Action:  "request.info.one",
		}
		reply, err := client.RequestRemoteAction(ctx, request)
		if err != nil {
			notify.StdMsgErr("could not contact " + rmt.ID)
		}
		rpcServices = []*cabal.Service{reply.GetServiceInfo()}
		can()
	}

	response := &cabal.ListReply{Status: "ack", Remotes: []*cabal.ProcessManager{remoteToRPC(rmt, rpcServices)}}
	return response, nil

}

func (c *controlServer) RestartAll(ctx context.Context, in *cabal.AllRequest) (*cabal.StatusReply, error) {
	rpcMessage("<- restart all request")

	// make sure that sender is gmbh-ctrl
	if in.GetSender() != "gmbh-ctrl" {
		rpcMessage("reporting invalid sender")
		return &cabal.StatusReply{Status: "invalid sender"}, nil
	}

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.StatusReply{Status: "internal server error"}, nil
	}

	pm.RestartAll()

	response := &cabal.StatusReply{Status: "ack"}
	return response, nil
}

func (c *controlServer) KillAll(ctx context.Context, in *cabal.AllRequest) (*cabal.StatusReply, error) {
	rpcMessage("<- restart all request")
	return &cabal.StatusReply{Status: "invalid"}, nil
}

func (c *controlServer) StopServer(ctx context.Context, in *cabal.StopRequest) (*cabal.StatusReply, error) {
	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.StatusReply{Status: "internal server error"}, nil
	}
	go func() {
		time.Sleep(time.Second * 2)
		pm.Shutdown(true)
	}()
	return &cabal.StatusReply{Status: "shutdown procedure started"}, nil
}

func (c *controlServer) ServerStatus(ctx context.Context, in *cabal.StatusRequest) (*cabal.StatusReply, error) {
	return &cabal.StatusReply{Status: "ack"}, nil
}

func (c *controlServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {

	rpcMessage("<- update service request; action=" + in.GetAction())
	rpcMessage(fmt.Sprintf("   details; sender=%s; target=%s; message=%s", in.GetSender(), in.GetTarget(), in.GetMessage()))

	// make sure that sender is gmbh-remote
	if in.GetSender() != "gmbh-remote" {
		rpcMessage("reporting invalid sender")
		return &cabal.ServiceUpdate{Message: "invalid sender"}, nil
	}

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.ServiceUpdate{Message: "internal server error"}, nil
	}

	if in.GetAction() == "remote.register" {
		id, address, err := pm.RegisterRemote()
		if err != nil {
			rpcMessage("Could not add; err=" + err.Error())
			return &cabal.ServiceUpdate{Message: "router error=" + err.Error()}, nil
		}
		update := &cabal.ServiceUpdate{
			Sender:  "core",
			Action:  "register",
			Target:  id,
			Status:  address,
			Message: "registered",
		}
		rpcMessage("sent registration response")
		return update, nil
	}

	if in.GetAction() == "shutdown.notification" {
		pm.MarkShutdown(in.GetMessage())
	}

	return &cabal.ServiceUpdate{Message: "invalid request"}, nil
}

func (c *controlServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	// rpcMessage("<- pong")

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
	}

	r, err := pm.LookupRemote(ping.GetFromID())
	if err != nil {
		rpcMessage("<- (nil)pong; could not find: " + ping.GetFromID())
		return nil, errors.New("")
	}
	r.UpdatePing(time.Now())

	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func remoteToRPC(r *RemoteServer, services []*cabal.Service) *cabal.ProcessManager {
	return &cabal.ProcessManager{
		ID:       r.ID,
		Address:  r.Address,
		Services: services,
	}
}

func rpcMessage(msg string) {
	notify.StdMsgMagentaNoPrompt("[rpc] " + msg)
}
