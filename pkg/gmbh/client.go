package gmbh

import (
	"context"
	"fmt"
	"time"

	"github.com/gimlet-gmbh/gimlet/gproto"
	"github.com/gimlet-gmbh/gimlet/ipc"
	"google.golang.org/grpc"
)

/*
 * rpc.go
 * Abe Dick
 * Nov 2018
 */

func getRPCClient() (gproto.CabalClient, error) {
	con, err := grpc.Dial("localhost:59999", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return gproto.NewCabalClient(con), nil

}

func getContextCancel() (context.Context, context.CancelFunc) {
	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return ctx, can
}

func makeRequest() (gproto.CabalClient, context.Context, context.CancelFunc, error) {
	client, err := getRPCClient()
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second)
	return client, ctx, can, nil
}

func _ephemeralRegisterService(name string, isClient bool, isServer bool) (string, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := gproto.RegServReq{
		NewServ: &gproto.NewService{
			Name:     name,
			Aliases:  []string{},
			IsClient: isClient,
			IsServer: isServer,
		},
	}

	reply, err := client.EphemeralRegisterService(ctx, &request)
	if err != nil {
		panic(err)
	}

	fmt.Println(reply.Status + "@" + reply.Address + "@" + reply.CorePath)

	return reply.Address, nil
}

func _makeDataRequest(target string, method string, data string) (ipc.Responder, error) {

	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	request := gproto.DataReq{
		Req: &gproto.Request{
			Sender: "test",
			Target: target,
			Method: method,
			Data1:  data,
		},
	}

	reply, err := client.MakeDataRequest(ctx, &request)
	if err != nil {
		// panic(err)
		fmt.Println(fmt.Errorf("%v", err.Error()))

		r := ipc.Responder{
			HadError:    true,
			ErrorString: err.Error(),
		}
		return r, err

	}

	return ipc.ResponderFromProto(*reply.Resp), nil
}

func _makeUnregisterRequest(name string) {
	client, ctx, can, err := makeRequest()
	if err != nil {
		panic(err)
	}
	defer can()

	_, _ = client.UnregisterService(ctx, &gproto.UnregisterReq{Name: name})
}
