package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
)

func listAll() {
	client, ctx, can, err := getClient(defaults.CONTROL_HOST + defaults.CONTROL_PORT)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	defer can()

	request := cabal.AllRequest{}
	reply, err := client.ListAll(ctx, &request)
	if err != nil {
		notify.StdMsgErr("error: " + err.Error())
	}
	pprintListAll(reply.Services)
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
		notify.StdMsgErr("error: " + err.Error())
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
		notify.StdMsgErr("error: " + err.Error())
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
	}
	notify.StdMsgBlue(reply.Status)
}
