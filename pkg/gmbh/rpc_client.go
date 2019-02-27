package gmbh

import (
	"errors"
	"strconv"
	"time"

	"github.com/gmbh-micro/rpc"
	"github.com/gmbh-micro/rpc/intrigue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

/**********************************************************************************
** RPCClient
**********************************************************************************/

func register(name string, isClient bool, isServer bool, mode string) (*registration, error) {

	client, ctx, can, err := rpc.GetCabalRequest(g.opts.standalone.CoreAddress, time.Second)
	if err != nil {
		return nil, errors.New("registration.gmbhUnavailable")
	}
	defer can()

	request := intrigue.NewServiceRequest{
		Service: &intrigue.NewService{
			Name:     name,
			Aliases:  []string{},
			IsClient: isClient,
			IsServer: isServer,
		},
	}

	reply, err := client.RegisterService(ctx, &request)
	if err != nil {
		if grpc.Code(err) == codes.Unavailable {
			return nil, errors.New("registration.gmbhUnavailable")
		}
		g.printer(grpc.Code(err).String())
		return nil, errors.New("registration.gmbhUnavailable")
	}

	if reply.Message == "acknowledged" {

		reg := reply.GetServiceInfo()

		r := &registration{
			id:          reg.GetID(),
			address:     reg.GetAddress(),
			fingerprint: reg.GetFingerprint(),
		}
		return r, nil
	}
	return nil, errors.New(reply.GetMessage())
}

func makeDataRequest(target, method string, data *Payload) (Responder, error) {

	client, ctx, can, err := rpc.GetCabalRequest(g.opts.standalone.CoreAddress, time.Second)
	if err != nil {
		return Responder{}, errors.New("data.gmbhUnavailable")
	}
	defer can()

	request := intrigue.DataRequest{
		Request: &intrigue.Request{
			Tport: &intrigue.Transport{
				Target: target,
				Method: method,
				Sender: g.opts.service.Name,
			},
			Pload: data.Proto(),
		},
	}

	mcs := strconv.Itoa(g.msgCounter)
	g.msgCounter++
	g.printer("<==" + mcs + "== target: " + target + ", method: " + method)

	reply, err := client.Data(ctx, &request)
	if err != nil {
		r := Responder{err: err.Error()}
		return r, err

	}
	g.printer(" ==" + mcs + "==> " + reply.String())
	if reply.Responder == nil {
		return responderFromProto(intrigue.Responder{}), nil
	}
	return responderFromProto(*reply.Responder), nil
}
