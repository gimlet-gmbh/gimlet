package main

import (
	"context"
	"time"

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
)

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

type controlServer struct{}

func (c *controlServer) StartService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {
	rpcMessage("<- Start; action=" + in.String())
	return &intrigue.Receipt{Error: "request.action.invalid"}, nil
}

func (c *controlServer) KillService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {
	rpcMessage("<- Kill; action=" + in.String())
	return &intrigue.Receipt{Error: "request.action.invalid"}, nil
}

func (c *controlServer) RestartService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {

	rpcMessage("<- Restart; action=" + in.String())

	request := in.GetRequest()
	remoteID := in.GetRemoteID()
	serviceID := in.GetTarget()

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}

	if request == "restart.all" {

		go pm.RestartAll()
		return &intrigue.Receipt{Message: "ack"}, nil
	} else if request == "restart.one" {

		remote, err := pm.LookupRemote(remoteID)
		if err != nil {
			rpcMessage("could not find remote")
			return &intrigue.Receipt{Error: "remote.notFound"}, nil
		}
		rpcMessage("found parent remote")
		pid := "-1"
		{
			client, ctx, can, err := rpc.GetRemoteRequest(remote.Address, time.Second*15)
			if err != nil {
				rpcMessage("could not contact " + remote.ID)
			}
			request := &intrigue.Action{
				Request: "service.restart.one",
				Target:  serviceID,
			}
			reply, err := client.NotifyAction(ctx, request)
			if err != nil {
				rpcMessage("could not contact " + remote.ID)
			}
			pid = reply.GetMessage()
			can()
		}
		rpcMessage("new pid=" + pid)
		return &intrigue.Receipt{Message: "pid=" + pid}, nil
	}

	return &intrigue.Receipt{Error: "request.action.unknown"}, nil

}

func (c *controlServer) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {

	rpcMessage("<- summary; action=" + in.String())

	request := in.GetRequest()
	remoteID := in.GetRemoteID()
	serviceID := in.GetTarget()

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &intrigue.SummaryReceipt{Error: "internal.pmref"}, nil
	}

	if request == "summary.all" {

		rpcrmts := []*intrigue.ProcessManager{}
		for _, re := range pm.GetAllRemotes() {
			{
				client, ctx, can, err := rpc.GetRemoteRequest(re.Address, time.Second*2)
				if err != nil {
					notify.LnBRedF("failed to contact\nid=%s; address=%s\nerr=%s", re.ID, re.Address, err.Error())
					continue
				}
				request := &intrigue.Action{
					Request: "request.info.all",
				}
				reply, err := client.Summary(ctx, request)
				if err != nil {
					notify.LnBRedF("failed to contact\nid=%s; address=%s\nerr=%s", re.ID, re.Address, err.Error())
					continue
				}

				rpcrmts = append(rpcrmts, reply.GetRemotes()...)
				can()
			}
		}

		return &intrigue.SummaryReceipt{
			Remotes: rpcrmts,
		}, nil

	} else if request == "summary.one" {

		rmt, err := pm.LookupRemote(remoteID)
		if err != nil {
			rpcMessage("could not find remote")
			return &intrigue.SummaryReceipt{Error: "remote.notFound"}, nil
		}

		rpcRemotes := []*intrigue.ProcessManager{}
		{
			client, ctx, can, err := rpc.GetRemoteRequest(rmt.Address, time.Second*5)
			if err != nil {
				// TODO add return here
				notify.StdMsgErr("could not contact " + rmt.ID)
			}
			request := &intrigue.Action{
				Target:  serviceID,
				Request: "request.info.one",
			}
			reply, err := client.Summary(ctx, request)
			if err != nil {
				// TODO add return here
				notify.StdMsgErr("could not contact " + rmt.ID)
			}
			rpcRemotes = append(rpcRemotes, reply.GetRemotes()...)
			can()
		}

		return &intrigue.SummaryReceipt{
			Remotes: rpcRemotes,
		}, nil
	}

	return &intrigue.SummaryReceipt{Error: "request.action.unknown"}, nil

}

func (c *controlServer) StopServer(ctx context.Context, in *intrigue.EmptyRequest) (*intrigue.Receipt, error) {

	rpcMessage("<- stop server request; action=" + in.String())

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}

	go func() {
		time.Sleep(time.Second * 2)
		pm.Shutdown(true)
	}()
	return &intrigue.Receipt{Message: "ack"}, nil

}

func (c *controlServer) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {

	rpcMessage("<- UpdateRegistration; serviceUpdate=" + in.String())

	request := in.GetRequest()
	message := in.GetMessage()

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}

	if request == "remote.register" {

		id, address, err := pm.RegisterRemote()
		if err != nil {
			rpcMessage("router.err=" + err.Error())
			return &intrigue.Receipt{Error: "router.err=" + err.Error()}, nil
		}

		rpcMessage("sent registration response")
		return &intrigue.Receipt{
			Message: "registered",
			ServiceInfo: &intrigue.ServiceSummary{
				Address: address,
				ID:      id,
			},
		}, nil

	} else if request == "shutdown.notif" {
		pm.MarkShutdown(message)
		return &intrigue.Receipt{
			Message: "ack",
		}, nil
	}

	return &intrigue.Receipt{Error: "request.action.unknown"}, nil
}

func (c *controlServer) Alive(ctx context.Context, ping *intrigue.Ping) (*intrigue.Pong, error) {
	rpcMessage("<- pong")

	fromID := ping.GetStatus()

	pm, err := GetProcM()
	if err != nil {
		rpcMessage("internal system error")
		return &intrigue.Pong{Error: "internal.pmref"}, nil
	}

	r, err := pm.LookupRemote(fromID)
	if err != nil {
		rpcMessage("<- (nil)pong; could not find: " + fromID)
		return &intrigue.Pong{Error: "not.found"}, nil
	}

	r.UpdatePing(time.Now())

	return &intrigue.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func remoteToRPC(r *RemoteServer, services []*intrigue.Service) *intrigue.ProcessManager {
	return &intrigue.ProcessManager{
		ID:       r.ID,
		Address:  r.Address,
		Services: services,
	}
}

func rpcMessage(msg string) {
	notify.StdMsgMagentaNoPrompt("[rpc] " + msg)
}
