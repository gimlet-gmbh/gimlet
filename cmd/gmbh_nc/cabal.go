package main

import (
	"context"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
)

func v(msg string) {
	notify.StdMsgBlueNoPrompt(" [cbl] " + msg)

}

var cnt int

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

	rv("-> Incoming Registration;")
	rv("   Name=%s; Aliases=%s;", in.GetNewServ().GetName(), in.GetNewServ().GetAliases())

	c, err := GetCore()
	if err != nil {
		return &cabal.RegServRep{Status: "internal error"}, nil
	}

	ns, err := c.Router.AddService(in.GetNewServ().GetName(), in.GetNewServ().GetAliases())
	if err != nil {
		return &cabal.RegServRep{Status: "error=" + err.Error()}, nil
	}

	reply := &cabal.RegServRep{
		Status: "acknowledged",

		Address:  ns.Address,
		ID:       ns.ID,
		CorePath: c.ProjectPath,
	}
	return reply, nil

}

func (s *cabalServer) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {
	t := time.Now()
	defer func() { cnt++ }()
	request := in.GetReq()
	rd("-%d-> Data request: %s", cnt, request.String())

	c, err := GetCore()
	if err != nil {
		rd("<-%d- could not get core error=%s", cnt, err.Error())
		return &cabal.DataResp{Status: &cabal.Status{Sender: "gmbh-core", Error: "internal server error"}}, nil
	}

	fwd, err := c.Router.LookupService(request.GetTarget())
	if err != nil {
		rd("<-%d- service not found error=%s", cnt, err.Error())
		return &cabal.DataResp{Status: &cabal.Status{Sender: "gmbh-core", Error: "service not found"}}, nil
	}

	client, ctx, can, err := rpc.GetCabalRequest(fwd.Address, time.Second*2)
	if err != nil {
		rd("<-%d- rpc error=%s", cnt, err.Error())
		return &cabal.DataResp{Status: &cabal.Status{Sender: "gmbh-core", Error: "internal server error"}}, nil
	}
	defer can()
	final, err := client.MakeDataRequest(ctx, in)
	if err != nil {
		rd("<-%d- could not forward error=%s", cnt, err.Error())
		return &cabal.DataResp{Status: &cabal.Status{Sender: "gmbh-core", Error: "could not forward"}}, nil
	}
	rd("<-%d-  elapsed time=%s", cnt, time.Since(t))
	return final, nil
}

func (s *cabalServer) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {

	rv("-> Unregister Request;")
	rv("   Name=%s; ID=%s; Address=%s", in.GetName(), in.GetId(), in.GetAddress())

	c, err := GetCore()
	if err != nil {
		return &cabal.UnregisterResp{
			Ack:    false,
			Status: &cabal.Status{Sender: "gmbh-core", Target: in.GetId(), Error: "internal server error"},
		}, nil
	}

	service, err := c.Router.LookupService(in.GetName())
	if err != nil {
		return &cabal.UnregisterResp{
			Ack:    false,
			Status: &cabal.Status{Sender: "gmbh-core", Target: in.GetId(), Error: "not found"},
		}, nil
	}

	service.UpdateState(Shutdown)

	return &cabal.UnregisterResp{
		Ack: true,
		Status: &cabal.Status{
			Sender: "gmbh-core",
			Target: in.GetId()},
	}, nil
}

func (s *cabalServer) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
	return &cabal.QueryResponse{}, nil
}

func (s *cabalServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {

	rv("-> Service Registration;")
	rv("   Sender=%s; Target=%s; Status=%s; Action=%s", in.GetSender(), in.GetTarget(), in.GetAction())
	rv("   Message=%s", in.GetMessage())

	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}

func (s *cabalServer) Alive(ctx context.Context, ping *cabal.Ping) (*cabal.Pong, error) {
	id := ping.GetID()
	c, err := GetCore()
	if err != nil {
		return &cabal.Pong{Status: &cabal.Status{Error: "internal server error"}}, nil
	}
	err = c.Router.Verify(id.GetName(), id.GetID(), id.GetAddress())
	if err != nil {
		rve("could not verify ping from id=%s; err=%s", id.GetID(), err.Error())
		return &cabal.Pong{Status: &cabal.Status{Sender: "core.NotVerified"}}, nil
	}
	return &cabal.Pong{Time: time.Now().Format(time.Stamp), Status: &cabal.Status{Sender: "core.verified"}}, nil
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
