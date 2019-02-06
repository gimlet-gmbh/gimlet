package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"

	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func handleErr(err error) string {
	if grpc.Code(err) == codes.Unavailable {
		return "could not connect to gmbhCore"
	}
	return "unsupported error code"
}

func listAll() {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{}
	reply, err := client.ListAll(ctx, &request)
	if err != nil {
		notify.StdMsgBlue("Could not contact gmbhServer")
		notify.StdMsgErr("error: "+err.Error(), 1)
		return
	}
	if reply.Length == 0 {
		notify.StdMsgBlue("no services to list")
	}
	pprintListAll(reply.Services)
}

func report() {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{}
	reply, err := client.ListAll(ctx, &request)
	if err != nil {
		notify.StdMsgBlue("Could not contact gmbhServer")
		notify.StdMsgErr("error: "+err.Error(), 1)
		return
	}
	if reply.Length == 0 {
		notify.StdMsgBlue("no services to list")
	}
	for _, s := range reply.GetServices() {
		pprintListOne(*s)
	}
}

func restartAll() {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{}
	reply, err := client.RestartAll(ctx, &request)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
		return
	}
	notify.StdMsgBlue(reply.GetStatus())
}

func listOne(id string) {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.SearchRequest{Id: id}
	reply, err := client.ListOne(ctx, &request)
	if err != nil {
		notify.StdMsgErr(handleErr(err))
		return
	}
	if reply.Length == 0 {
		notify.StdMsgErr("could not find service with id: " + id)
		return
	}
	pprintListOne(*reply.Services[0])
}

func restartOne(id string) {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.SearchRequest{Id: id}
	reply, err := client.RestartService(ctx, &request)
	if err != nil {
		fmt.Println(err)
		notify.StdMsgErr("error: " + err.Error())
		return
	}

	notify.StdMsgBlue(reply.GetStatus())
}

func shutdown() {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
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
