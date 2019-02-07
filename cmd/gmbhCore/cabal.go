package main

/*
 * cabal.go
 * Implements the gRPC server and client for the gmbhCore Cabal Server
 * Abe Dick
 * Nov 2018
 */

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"google.golang.org/grpc"
)

/////////////////////////////////////////////////////////////////////////
// CLIENT
/////////////////////////////////////////////////////////////////////////

func makeCabalRequest(address string) (cabal.CabalClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return cabal.NewCabalClient(con), ctx, can, nil
}

func requestQueryData(address string) (*cabal.QueryResponse, error) {
	client, ctx, can, err := makeCabalRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	req := cabal.QueryRequest{
		Query: cabal.QueryRequest_STATUS,
	}
	return client.QueryStatus(ctx, &req)
}

/////////////////////////////////////////////////////////////////////////
// SERVER
/////////////////////////////////////////////////////////////////////////

// cabalServer is for gRPC interface for the gmbhCore service coms server
type cabalServer struct{}

func (s *cabalServer) EphemeralRegisterService(ctx context.Context, in *cabal.RegServReq) (*cabal.RegServRep, error) {

	lookupService, err := core.Router.LookupService(in.NewServ.GetName())
	if err != nil {
		if err.Error() == "router.LookupService.nameNotFound" {
			if in.NewServ.GetMode() == cabal.NewService_REMOTE {
				lookupService, err = core.registerRemoteService(
					in.GetNewServ().GetName(),
					in.GetNewServ().GetAliases(),
					in.GetNewServ().GetIsClient(),
					in.GetNewServ().GetIsServer())
				if err != nil {
					return &cabal.RegServRep{Status: err.Error()}, nil
				}
			}
		}
	}

	if !core.Config.Daemon {
		notify.StdMsgMagenta(fmt.Sprintf("<(%s)- processing ephem-reg request; name=(%s); aliases=(%s); mode=(%s)", lookupService.ID, in.NewServ.GetName(), strings.Join(in.NewServ.GetAliases(), ","), lookupService.GetMode()))
		if lookupService.Static.IsServer {
			notify.StdMsgMagenta(fmt.Sprintf("-(%s)> success; address=(%v)", lookupService.ID, lookupService.Address))
		} else {
			notify.StdMsgMagenta(fmt.Sprintf("-(%s)> success;", lookupService.ID))
		}
	}

	reply := &cabal.RegServRep{
		Status:   "acknowledged",
		ID:       lookupService.ID,
		CorePath: core.ProjectPath,
		Address:  lookupService.Address,
	}

	return reply, nil
}

func (s *cabalServer) MakeDataRequest(ctx context.Context, in *cabal.DataReq) (*cabal.DataResp, error) {

	reqHandler := newRequestHandler(in.GetReq())
	reqHandler.Fulfill()

	return &cabal.DataResp{Resp: reqHandler.GetResponder()}, nil
}

func (s *cabalServer) UnregisterService(ctx context.Context, in *cabal.UnregisterReq) (*cabal.UnregisterResp, error) {
	reply := &cabal.UnregisterResp{Awk: false}
	return reply, nil
}

func (s *cabalServer) QueryStatus(ctx context.Context, in *cabal.QueryRequest) (*cabal.QueryResponse, error) {
	return nil, nil
}

func (s *cabalServer) UpdateServiceRegistration(ctx context.Context, in *cabal.ServiceUpdate) (*cabal.ServiceUpdate, error) {
	return &cabal.ServiceUpdate{Message: "unimp"}, nil
}
