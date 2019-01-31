package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
)

func listAll() {
	client, ctx, can, err := getClient(CTRLSERVER)
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
func shutdown() {
	client, ctx, can, err := getClient(CTRLSERVER)
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
