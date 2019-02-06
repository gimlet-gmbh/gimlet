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
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
	"github.com/rs/xid"
	"google.golang.org/grpc"
)

/////////////////////////////////////////////////////////////////////////
// CLIENT
/////////////////////////////////////////////////////////////////////////

func getRPCClient(address string) (cabal.CabalClient, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return cabal.NewCabalClient(con), nil
}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest(address string) (cabal.CabalClient, context.Context, context.CancelFunc, error) {
	client, err := getRPCClient(address)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return client, ctx, can, nil
}

func requestQueryData(address string) (*cabal.QueryResponse, error) {
	client, ctx, can, err := makeRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	req := cabal.QueryRequest{
		Query: cabal.QueryRequest_STATUS,
	}

	// reply, err := client.QueryStatus(ctx, &req)
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

	if !core.Config.Daemon && core.Config.Verbose {
		notify.StdMsgLog(fmt.Sprintf("<- Ephemeral Registration Request: %s", in.NewServ.GetName()))
		if lookupService.Static.IsServer {
			notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged with address: %v", in.NewServ.GetName(), lookupService.Address))
		} else {
			notify.StdMsgLog(fmt.Sprintf("-> %s: acknowledged", in.NewServ.GetName()))
		}
	} else {
		// log this somewhere
	}

	reply := &cabal.RegServRep{
		Status:   "acknowledged",
		ID:       xid.New().String(),
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
