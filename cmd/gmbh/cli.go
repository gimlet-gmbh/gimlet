package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func listAll() {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{Sender: "gmbh-ctrl"}
	reply, err := client.ListAll(ctx, &request)
	if err != nil {
		notify.StdMsgBlue("Could not contact gmbhServer")
		notify.StdMsgErr("error: "+err.Error(), 1)
		return
	}
	pprintListAll(reply.GetRemotes())
}

func runReport() {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{Sender: "gmbh-ctrl"}
	reply, err := client.ListAll(ctx, &request)
	if err != nil {
		notify.StdMsgBlue("Could not contact gmbhServer")
		notify.StdMsgErr("error: "+err.Error(), 1)
		return
	}
	if reply.Length == 0 {
		notify.StdMsgBlue("no services to list")
	}

	pprintListOne(reply.GetRemotes())
}

func restartAll() {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{Sender: "gmbh-ctrl"}
	reply, err := client.RestartAll(ctx, &request)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
		return
	}
	notify.StdMsgBlue(reply.GetStatus())
}

func listOne(id string) {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second*5)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	splitID := strings.Split(id, "-")
	if len(splitID) != 2 {
		notify.StdMsgErr("could not parse id")
		return
	}

	request := cabal.SearchRequest{Sender: "gmbh-ctrl", ParentID: splitID[0], Id: splitID[1]}
	reply, err := client.ListOne(ctx, &request)
	if err != nil {
		notify.StdMsgErr(handleErr(err))
		return
	}

	if reply.GetStatus() != "ack" {
		notify.StdMsgErr("could not find service with id: " + id)
		notify.StdMsgErr("report from core=" + reply.GetStatus())
		return
	}
	pprintListOne(reply.GetRemotes())
}
func restartOne(id string) {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second*20)
	if err != nil {
		notify.StdMsgErr("client error: " + err.Error())
	}
	defer can()

	splitID := strings.Split(id, "-")
	if len(splitID) != 2 {
		notify.StdMsgErr("could not parse id")
		return
	}

	request := cabal.SearchRequest{Sender: "gmbh-ctrl", ParentID: splitID[0], Id: splitID[1]}
	reply, err := client.RestartService(ctx, &request)
	if err != nil {
		fmt.Println(err)
		notify.StdMsgErr("send error: " + err.Error())
		return
	}

	notify.StdMsgBlue(reply.GetStatus())
}

func shutdown() {
	client, ctx, can, err := rpc.GetControlRequest(defaults.CONTROL_HOST+defaults.CONTROL_PORT, time.Second)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.StopRequest{}
	reply, err := client.StopServer(ctx, &request)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
		return
	}
	notify.StdMsgBlue(reply.Status)
}

func handleErr(err error) string {
	if grpc.Code(err) == codes.Unavailable {
		return "could not connect to gmbhCore"
	}
	return "unsupported error code"
}
