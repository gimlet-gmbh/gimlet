package main

/*
 * main.go (gmbhContainer)
 * Abe Dick
 * February 2019
 */

import "github.com/gmbh-micro/notify"

const PATH_TO_CONFIG = "/config/gmbh.yaml"

func main() {
	notify.SetTag("[gmbh-exp] ")
	notify.StdMsgBlue("gmbh container process manager v0.0.1")
}
