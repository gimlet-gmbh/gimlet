package main

import (
	"context"
	"strings"
	"time"

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"google.golang.org/grpc/metadata"
)

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

type controlServer struct{}

func (c *controlServer) StartService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {
	rpcprint("<- Start; action=" + in.String())
	return &intrigue.Receipt{Error: "request.action.invalid"}, nil
}

func (c *controlServer) KillService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {
	rpcprint("<- Kill; action=" + in.String())
	return &intrigue.Receipt{Error: "request.action.invalid"}, nil
}

func (c *controlServer) RestartService(ctx context.Context, in *intrigue.Action) (*intrigue.Receipt, error) {

	rpcprint("<- Restart; action=" + in.String())

	request := in.GetRequest()
	remoteID := in.GetRemoteID()
	serviceID := in.GetTarget()

	pm, err := GetProcM()
	if err != nil {
		rpcprint("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}

	if request == "restart.all" {

		go pm.RestartAll()
		return &intrigue.Receipt{Message: "ack"}, nil
	} else if request == "restart.one" {

		remote, err := pm.LookupRemote(remoteID)
		if err != nil {
			rpcperr("could not find remote")
			return &intrigue.Receipt{Error: "remote.notFound"}, nil
		}
		rpcprint("found parent remote")
		pid := "-1"
		{
			client, ctx, can, err := rpc.GetRemoteRequest(remote.Address, time.Second*15)
			if err != nil {
				rpcperr("could not contact " + remote.ID)
			}
			request := &intrigue.Action{
				Request: "service.restart.one",
				Target:  serviceID,
			}
			reply, err := client.NotifyAction(ctx, request)
			if err != nil {
				rpcperr("could not contact " + remote.ID)
			}
			pid = reply.GetMessage()
			can()
		}
		rpcprint("new pid=" + pid)
		return &intrigue.Receipt{Message: "pid=" + pid}, nil
	}

	return &intrigue.Receipt{Error: "request.action.unknown"}, nil

}

func (c *controlServer) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {

	rpcprint("<- summary; action=" + in.String())

	request := in.GetRequest()
	remoteID := in.GetRemoteID()
	serviceID := in.GetTarget()

	pm, err := GetProcM()
	if err != nil {
		rpcperr("internal system error")
		return &intrigue.SummaryReceipt{Error: "internal.pmref"}, nil
	}

	if request == "summary.all" {

		rpcrmts := []*intrigue.ProcessManager{}
		for _, re := range pm.GetAllRemotes() {
			{
				client, ctx, can, err := rpc.GetRemoteRequest(re.Address, time.Second*2)
				if err != nil {
					rpcperr("failed to contact\nid=%s; address=%s\nerr=%s", re.ID, re.Address, err.Error())
					continue
				}
				request := &intrigue.Action{
					Request: "request.info.all",
				}
				reply, err := client.Summary(ctx, request)
				if err != nil {
					rpcperr("failed to contact\nid=%s; address=%s\nerr=%s", re.ID, re.Address, err.Error())
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
			rpcperr("could not find remote")
			return &intrigue.SummaryReceipt{Error: "remote.notFound"}, nil
		}

		rpcRemotes := []*intrigue.ProcessManager{}
		{
			client, ctx, can, err := rpc.GetRemoteRequest(rmt.Address, time.Second*5)
			if err != nil {
				// TODO add return here
				rpcperr("could not contact " + rmt.ID)
			}
			request := &intrigue.Action{
				Target:  serviceID,
				Request: "request.info.one",
			}
			reply, err := client.Summary(ctx, request)
			if err != nil {
				// TODO add return here
				rpcperr("could not contact " + rmt.ID)
			}
			rpcRemotes = append(rpcRemotes, reply.GetRemotes()...)
			can()
		}

		return &intrigue.SummaryReceipt{
			Remotes: rpcRemotes,
		}, nil
	} else if request == "summary.reflect" {

	}

	return &intrigue.SummaryReceipt{Error: "request.action.unknown"}, nil

}

func (c *controlServer) StopServer(ctx context.Context, in *intrigue.EmptyRequest) (*intrigue.Receipt, error) {

	rpcprint("<- stop server request; action=" + in.String())

	pm, err := GetProcM()
	if err != nil {
		rpcperr("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}
	go func() {
		time.Sleep(time.Second * 2)
		pm.Shutdown(true)
	}()
	return &intrigue.Receipt{Message: "ack"}, nil

}

func (c *controlServer) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {

	rpcprint("<- UpdateRegistration; serviceUpdate=" + in.String())

	request := in.GetRequest()
	message := in.GetMessage()

	pm, err := GetProcM()
	if err != nil {
		rpcperr("internal system error")
		return &intrigue.Receipt{Error: "internal.pmref"}, nil
	}

	if request == "remote.register" {

		// message is mode
		id, address, fingerprint, err := pm.RegisterRemote(message)
		if err != nil {
			rpcperr("router.err=" + err.Error())
			return &intrigue.Receipt{Error: "router.err=" + err.Error()}, nil
		}

		rpcprint("sent registration response")
		return &intrigue.Receipt{
			Message: "registered",
			ServiceInfo: &intrigue.ServiceSummary{
				Address:     address,
				ID:          id,
				Fingerprint: fingerprint,
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
	// rpcMessage("<- pong")

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		rpcperr("Could not get metadata")
	}

	pm, err := GetProcM()
	if err != nil {
		rpcperr("internal system error")
		return &intrigue.Pong{Error: "internal.pmref"}, nil
	}

	id := strings.Join(md.Get("sender"), "")
	fp := strings.Join(md.Get("fingerprint"), "")

	verified := pm.Verify(id, fp)
	if !verified {
		rpcperr("<- (nil)pong; could not verify: " + id)
		return &intrigue.Pong{Error: "verification.error"}, nil
	}

	return &intrigue.Pong{Time: time.Now().Format(time.Stamp)}, nil
}

func remoteToRPC(r *RemoteServer, services []*intrigue.Service) *intrigue.ProcessManager {
	return &intrigue.ProcessManager{
		ID:       r.ID,
		Address:  r.Address,
		Services: services,
	}
}

func rpcprint(format string, a ...interface{}) {
	notify.LnMagentaF("[control] "+format, a...)

}

func rpcperr(format string, a ...interface{}) {
	notify.LnRedF("[control] "+format, a...)
}
