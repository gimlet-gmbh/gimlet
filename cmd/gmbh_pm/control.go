package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
)

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

type controlServer struct{}

func (c *controlServer) StartService(ctx context.Context, in *cabal.StartRequest) (*cabal.StartReply, error) {
	// TODO: Implement
	return nil, nil
}

func (c *controlServer) RestartService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	// cc, err := getCore()
	// if err != nil {
	// 	return nil, errors.New("gmbh system error, could not locate instance of core")
	// }

	// target, err := cc.Router.LookupServiceID(in.GetId())
	// if err != nil {
	// 	return &cabal.StatusReply{Status: "could not find service: " + err.Error()}, nil
	// }
	// if target.Mode == service.Managed {
	// 	pid, err := target.Process.Restart(false)
	// 	if err != nil {
	// 		return &cabal.StatusReply{Status: "could not restart service: " + err.Error()}, nil
	// 	}
	// 	return &cabal.StatusReply{Status: "pid=" + strconv.Itoa(pid)}, nil
	// }
	// pid, err := target.Restart()
	// if err != nil {
	// 	return &cabal.StatusReply{Status: "issue restarting service"}, nil
	// }
	// return &cabal.StatusReply{Status: "pid=" + pid}, nil
	return nil, nil
}

func (c *controlServer) KillService(ctx context.Context, in *cabal.SearchRequest) (*cabal.StatusReply, error) {
	// TODO: Implement
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
		rpcRemote := &cabal.ProcessManager{
			ID:      re.ID,
			Address: re.Address,
		}
		rpcrmts = append(rpcrmts, rpcRemote)
	}

	// TODO: Need to ping and get process information from each remote

	response := &cabal.ListReply{
		Status: "ack",
		Remote: rpcrmts,
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

	rmt, err := pm.LookupRemote(in.GetId())
	if err != nil {
		return &cabal.ListReply{Status: "Not Found"}, nil
	}

	rpcrmt := &cabal.ProcessManager{
		ID:      rmt.ID,
		Address: rmt.Address,
	}

	// TODO: Need to ping and get process information from each remote

	response := &cabal.ListReply{Status: "ack", Remote: []*cabal.ProcessManager{rpcrmt}}
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
	pm.Shutdown(true)
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

	return &cabal.ServiceUpdate{Message: "invalid request"}, nil
}

func (c *controlServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	rpcMessage("<- pong")
	return &cabal.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func rpcMessage(msg string) {
	notify.StdMsgMagentaNoPrompt("[rpc] " + msg)
}
