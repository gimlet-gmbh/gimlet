package core

/*
 * rpcInternalClient.go
 * Abe Dick
 * Nov 2018
 */

import (
	"context"
	"time"

	"github.com/gimlet-gmbh/gimlet/cabal"
	"github.com/gimlet-gmbh/gimlet/notify"
	"google.golang.org/grpc"
)

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

func _makeDataRequest(req cabal.Request, address string) (*cabal.Responder, error) {
	client, ctx, can, err := makeRequest(address)
	if err != nil {
		panic(err)
	}
	defer can()

	request := cabal.DataReq{
		Req: &req,
	}

	reply, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		notify.StdMsgErr(err.Error(), 1)
		r := cabal.Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return &r, nil
	}

	return reply.Resp, nil
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
