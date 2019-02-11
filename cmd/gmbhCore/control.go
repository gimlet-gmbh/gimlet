package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/service"
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

	cc.Router.Reconcile()

	rpcManaged := []*cabal.Service{}
	rpcPlanetary := []*cabal.Service{}
	rpcRemote := []*cabal.ProcessManager{}

	for _, s := range cc.Router.GetAllServices() {
		if s.Mode == service.Managed {
			rpcManaged = append(rpcManaged, ServiceToRPC(*s))
		} else if s.Mode == service.Remote {

			pm, err := cc.Router.LookupProcessManager(s.Static.Name)
			if err != nil {
				epm := &cabal.ProcessManager{
					Services: []*cabal.Service{
						&cabal.Service{
							Name:   s.Static.Name,
							Mode:   "Remote",
							Errors: []string{"could not contact"},
						},
					},
				}
				rpcRemote = append(rpcRemote, epm)
				continue
			}

			client, con, can, err := rpc.GetRemoteRequest(pm.Address, time.Second)

			request := &cabal.Action{
				Action: "request.info",
			}

			reply, err := client.RequestRemoteAction(con, request)
			if err != nil {
				epm := &cabal.ProcessManager{
					Services: []*cabal.Service{
						&cabal.Service{
							Name:   s.Static.Name,
							Mode:   "Remote",
							Errors: []string{"could not contact"},
						},
					},
				}
				rpcRemote = append(rpcRemote, epm)
				continue
			}

			pmData := &cabal.ProcessManager{
				ID:       s.Parent.ID,
				Name:     s.Parent.Name,
				Address:  s.Parent.Address,
				Services: []*cabal.Service{reply.GetServiceInfo()},
			}
			rpcRemote = append(rpcRemote, pmData)

			can()
		} else {
			ns := &cabal.Service{
				Name:    s.Static.Name,
				Id:      s.ID,
				Address: s.Address,
				Mode:    "Planetary",
			}
			rpcPlanetary = append(rpcPlanetary, ns)
		}
	}

	reply := cabal.ListReply{
		Managed:   rpcManaged,
		Remote:    rpcRemote,
		Planetary: rpcPlanetary,
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
		Length:  1,
		Managed: []*cabal.Service{ServiceToRPC(*target)},
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

func (c *controlServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
	notify.StdMsgCyanNoPrompt(fmt.Sprintf("[ pm ] <- Update Service Request; sender=(%s); target=(%s); action=(%s); message=(%s);", in.GetSender(), in.GetTarget(), in.GetAction(), in.GetMessage()))

	if in.GetSender() != "gmbh-container" {
		return &cabal.ServiceUpdate{Message: "invalid sender"}, nil
	}

	cc, err := getCore()
	if err != nil {
		return nil, errors.New("gmbh system error, could not locate instance of core")
	}

	if in.GetAction() == "container.register" {
		c, err := cc.Router.AddProcessManager(in.GetMessage())
		if err != nil {
			return &cabal.ServiceUpdate{Message: err.Error()}, nil
		}
		return &cabal.ServiceUpdate{
			Target:  in.GetSender(),
			Sender:  "core",
			Message: "added process manager",
			Action:  c.Address,
			Status:  c.ID,
		}, nil
	}

	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

/////////////////////////////////////////////////////////////////////////
// CLIENT
/////////////////////////////////////////////////////////////////////////
