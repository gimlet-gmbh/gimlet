package main

import (
	"context"
	"strings"
	"time"

	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"google.golang.org/grpc/metadata"
)

var cnt int

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) RegisterService(ctx context.Context, in *intrigue.NewServiceRequest) (*intrigue.Receipt, error) {

	print("-> Incoming registration; Request=%s", in.String())

	c, err := GetCore()
	if err != nil {
		return &intrigue.Receipt{Error: "error.coreref"}, nil
	}

	newService := in.GetService()

	ns, err := c.Router.AddService(newService.GetName(), newService.GetAliases(), newService.GetPeerGroups(), in.GetEnv(), in.GetAddress())
	if err != nil {
		return &intrigue.Receipt{Error: err.Error()}, nil
	}

	return &intrigue.Receipt{
		Message: "acknowledged",
		ServiceInfo: &intrigue.ServiceSummary{
			Address:     ns.Address,
			ID:          ns.ID,
			Fingerprint: ns.Fingerprint,
		},
	}, nil

}

func (s *cabalServer) UpdateRegistration(ctx context.Context, in *intrigue.ServiceUpdate) (*intrigue.Receipt, error) {

	print("-> Update Registration; Update=%s", in.String())

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

	tport := in.GetRequest().GetTport()
	print("-%d-> Data request: %s", cnt, tport.String())

	c, err := GetCore()
	if err != nil {
		print("<-%d- could not get core error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "core.ref"}, nil
	}

	fwd, err := c.Router.LookupService(tport.GetTarget())
	if err != nil {
		print("<-%d- service not found error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "service.notFound"}, nil
	}

	client, ctx, can, err := rpc.GetCabalRequest(fwd.Address, time.Second*2)
	if err != nil {
		print("<-%d- rpc error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "rpc error=" + err.Error()}, nil
	}
	defer can()
	final, err := client.Data(ctx, in)
	if err != nil {
		print("<-%d- could not forward error=%s", cnt, err.Error())
		return &intrigue.DataResponse{Error: "unableToForward"}, nil
	}
	print("<-%d- elapsed time=%s", cnt, time.Since(t))
	return final, nil
}

func (s *cabalServer) WhoIs(ctx context.Context, in *intrigue.WhoIsRequest) (*intrigue.WhoIsResponse, error) {

	print("-> WhoIsRequest=%s", in.String())
	target := in.GetTarget()
	sender := in.GetSender()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &intrigue.WhoIsResponse{Error: "invalid request"}, nil
	}

	c, err := GetCore()
	if err != nil {
		return &intrigue.WhoIsResponse{Error: "core.ref"}, nil
	}

	name := strings.Join(md.Get("sender"), "")
	fp := strings.Join(md.Get("fingerprint"), "")

	verified := c.Router.Verify(name, fp)
	if verified != nil {
		print("could not verify %s; err=%s", name, verified.Error())
		return &intrigue.WhoIsResponse{Error: verified.Error()}, nil
	}

	addr, err := c.Router.GrantPermissions(sender, target)
	if err != nil {
		if err.Error() == "denied" {
			print("<- mismatch peer groups; %s -> %s", sender, target)
			return &intrigue.WhoIsResponse{Error: "permission.denied"}, nil
		}
		print("<- peer group err=%s; %s -> %s", err.Error(), sender, target)
		return &intrigue.WhoIsResponse{Error: "server.error"}, nil
	}

	print("<- granted; %s -> %s", sender, target)
	return &intrigue.WhoIsResponse{TargetAddress: addr}, nil

}

func (s *cabalServer) Summary(ctx context.Context, in *intrigue.Action) (*intrigue.SummaryReceipt, error) {

	print("-> Update Registration; Update=%s", in.String())

	c, err := GetCore()
	if err != nil {
		print("could not get core error=%s", cnt, err.Error())
		return &intrigue.SummaryReceipt{Error: "core.ref"}, nil
	}

	// add core itself
	ccs := &intrigue.CoreService{
		Name:     "CoreData",
		Address:  c.conf.Address,
		ParentID: c.parentID,
	}

	request := in.GetRequest()
	if request == "request.info.all" {
		return &intrigue.SummaryReceipt{
			Services: c.Router.GetCoreServiceData(ccs),
			Error:    "",
		}, nil
	}

	return &intrigue.SummaryReceipt{Error: "unimp"}, nil
}

func (s *cabalServer) Alive(ctx context.Context, ping *intrigue.Ping) (*intrigue.Pong, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &intrigue.Pong{Error: "invalid request"}, nil
	}

	c, err := GetCore()
	if err != nil {
		return &intrigue.Pong{Error: "core.ref"}, nil
	}

	name := strings.Join(md.Get("sender"), "")
	fp := strings.Join(md.Get("fingerprint"), "")

	verified := c.Router.Verify(name, fp)
	if verified != nil {
		print("could not verify %s; err=%s", name, verified.Error())
		return &intrigue.Pong{Error: verified.Error()}, nil
	}

	return &intrigue.Pong{Time: time.Now().Format(time.Stamp), Status: "core.verified"}, nil
}
