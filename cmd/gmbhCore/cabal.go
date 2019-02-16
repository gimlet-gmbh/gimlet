package main

import (
	"context"
	"time"

	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
)

func v(msg string) {
	notify.StdMsgBlueNoPrompt(" [cbl] " + msg)

}

var cnt int

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) RegisterService(ctx context.Context, in *intrigue.NewServiceRequest) (*intrigue.Receipt, error) {

	rv("-> Incoming registration; Request=%s", in.String())

	c, err := GetCore()
	if err != nil {
		return &intrigue.Receipt{Error: "error.coreref"}, nil
	}

	newService := in.GetService()

	ns, err := c.Router.AddService(newService.GetName(), newService.GetAliases())
	if err != nil {
		return &intrigue.Receipt{Error: err.Error()}, nil
	}

	return &intrigue.Receipt{
		Message: "acknowledged",
		ServiceInfo: &intrigue.ServiceSummary{
			Address: ns.Address,
			ID:      ns.ID,
		},
	}, nil

}

func (s *cabalServer) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {

	rv("-> Update Registration; Update=%s", in.String())

	request := in.GetRequest()
	name := in.GetMessage()

	c, err := GetCore()
	if err != nil {
		return &intrigue.Receipt{Error: "error.coreref"}, nil
	}

	if request == "shutdown.notif" {
		service, err := c.Router.LookupService(name)
		if err != nil {
			return &intrigue.Receipt{
				Error: "service.notFound",
			}, nil
		}

		service.UpdateState(Shutdown)
		return &intrigue.Receipt{Message: "ack"}, nil
	}

	return &intrigue.Receipt{Error: "invalid request"}, nil
}

func (s *cabalServer) Data(ctx context.Context, in *intrigue.DataRequest) (*intrigue.DataResponse, error) {

	t := time.Now()
	defer func() { cnt++ }()

	request := in.GetRequest()
	rd("-%d-> Data request: %s", cnt, request.String())

	c, err := GetCore()
	if err != nil {
		rd("<-%d- could not get core error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "core.ref"}, nil
	}

	fwd, err := c.Router.LookupService(request.GetTarget())
	if err != nil {
		rd("<-%d- service not found error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "service.notFound"}, nil
	}

	client, ctx, can, err := rpc.GetCabalRequest(fwd.Address, time.Second*2)
	if err != nil {
		rd("<-%d- rpc error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "rpc error=" + err.Error()}, nil
	}
	defer can()
	final, err := client.Data(ctx, in)
	if err != nil {
		rd("<-%d- could not forward error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "unableToForward"}, nil
	}
	rd("<-%d-  elapsed time=%s", cnt, time.Since(t))
	return final, nil
}

func (s *cabalServer) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {
	return &intrigue.SummaryReceipt{Error: "unimp"}, nil
}

func (s *cabalServer) Alive(ctx context.Context, ping *intrigue.Ping) (*intrigue.Pong, error) {

	// id := ping.GetStatus()

	// c, err := GetCore()
	// if err != nil {
	// 	return &intrigue.Pong{Error: "core.ref"}, nil
	// }

	// TODO Need to verify ping info in meta data
	// err = c.Router.Verify(id.GetName(), id.GetID(), id.GetAddress())
	// if err != nil {
	// 	rve("could not verify ping from id=%s; err=%s", id.GetID(), err.Error())
	// 	return &cabal.Pong{Status: &cabal.Status{Sender: "core.NotVerified"}}, nil
	// }

	return &intrigue.Pong{Time: time.Now().Format(time.Stamp), Status: "core.verified"}, nil
}

func rv(msg string, a ...interface{}) {
	notify.LnMagentaF("[rpc] "+msg, a...)
}

func rd(msg string, a ...interface{}) {
	notify.LnCyanF("[data] "+msg, a...)
}

func rve(msg string, a ...interface{}) {
	notify.LnRedF("[rpc] "+msg, a...)
}
