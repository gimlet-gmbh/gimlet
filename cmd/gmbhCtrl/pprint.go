package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gmbh-micro/cabal"
	"github.com/gmbh-micro/notify"
)

func pprintListOne(service cabal.Service) {
	notify.StdMsg("Displaying service information")
	notify.StdMsgNoPrompt(" ID:\t"+service.GetId(), 2)
	notify.StdMsgNoPrompt(" Name:\t"+service.GetName(), 2)
	notify.StdMsgNoPrompt(" PID:\t"+strconv.Itoa(int(service.GetPid())), 2)
	notify.StdMsgNoPrompt(" Start:\t"+service.GetStartTime(), 2)
	if service.GetStatus() == "Failed" {
		notify.StdMsgErrNoPrompt(" Status:\t"+service.GetStatus(), 2)
		notify.StdMsgErrNoPrompt(" Failed:\t"+service.GetFailTime(), 2)
	} else if service.GetStatus() == "degraded" {
		notify.StdMsgErrNoPrompt(" Status:\t"+service.GetStatus(), 2)
	} else {
		notify.StdMsgGreenNoPrompt(" Status:\t"+service.GetStatus(), 2)
	}
	notify.StdMsgNoPrompt(" Fails:\t"+strconv.Itoa(int(service.GetFails())), 2)
	notify.StdMsgNoPrompt(" Restarts:"+strconv.Itoa(int(service.GetRestarts())), 2)
	notify.StdMsgNoPrompt(" Path:\t"+service.GetPath(), 2)
	notify.StdMsgNoPrompt(" Logs:\t"+service.GetLogPath(), 2)
	errs := service.GetErrors()
	if len(errs) <= 1 {
		notify.StdMsgNoPrompt(" Errors:\t"+strings.Join(errs, ","), 2)
	} else {
		notify.StdMsgNoPrompt(" Errors:\t"+errs[0], 2)
		errs = errs[1:]
		for _, e := range errs {
			notify.StdMsgNoPrompt("        \t"+e, 2)
		}
	}
}

// Don't look too hard at this...
func pprintListAll(processes []*cabal.Service) {

	nameLength, largestID, _ := getDetails(processes)
	color.Set(color.Underline)

	fmt.Printf(" ID  ")
	if largestID >= 10 && largestID < 100 {
		fmt.Printf(" ")
	} else if largestID >= 100 {
		fmt.Printf("  ")
	}
	fmt.Printf("\u2502 PID   \u2502 Status      \u2502 Uptime \u2502 Name")
	for i := 0; i < nameLength-2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("| Path        \n")
	color.Unset()

	for _, p := range processes {

		id, _ := strconv.Atoi(p.Id)
		if (largestID >= 10 && largestID < 100) && id < 10 {
			fmt.Printf(" ")
		} else if largestID >= 100 && id < 100 {
			fmt.Printf("  ")
		}

		fmt.Printf(" %s   \u2502 ", p.Id)
		if p.Pid == -1 {
			fmt.Printf("----- \u2502 ")
		} else {
			if p.Pid < 10000 {
				fmt.Printf(" ")
			}
			fmt.Printf("%d \u2502 ", p.Pid)
		}

		extraSpace := false
		sColor := color.FgGreen
		if p.Status != "Running" {
			sColor = color.FgRed
			extraSpace = true
		}

		color.Set(sColor)
		fmt.Printf("[ %s ]", p.Status)
		color.Unset()
		if extraSpace {
			fmt.Printf(" ")
		}

		t, err := time.Parse(time.RFC3339, p.StartTime)
		if err != nil {
			fmt.Printf(" |       ")
		} else {

			dur := time.Now().Sub(t).Truncate(time.Minute)
			var durStr string
			if dur >= time.Minute {

				durStr = shortDur(dur)
				fmt.Printf(" | %v", durStr)

				if dur < (time.Minute * 10) {
					fmt.Printf("    ")
				} else if dur < (time.Hour) {
					fmt.Printf("   ")
				} else if dur < (time.Hour * 10) {
					fmt.Printf("  ")
				}

			} else {
				fmt.Printf(" | <1m   ")
			}

		}

		fmt.Printf(" ")

		fmt.Printf("\u2502 ")
		color.Set(color.FgBlue)
		fmt.Printf("%s", p.Name)
		color.Unset()
		if false {
			fmt.Printf(" \u2502 %s", p.Path)
		}

		for i := 0; i < (nameLength - len(p.Name) + 2); i++ {
			fmt.Printf(" ")
		}

		fmt.Printf("\u2502 ")
		fmt.Printf("%v", p.Path)
		fmt.Printf("\n")
	}
}

func shortDur(d time.Duration) string {
	str := d.String()
	e := str[:len(str)-2]
	return e
}

// Longest name length, largest id
func getDetails(processes []*cabal.Service) (longestName int, longestID int, longestPath int) {
	ln := 0
	li := 0
	lp := 0
	for _, p := range processes {
		if len(p.Name) > ln {
			ln = len(p.Name)
		}
		id, _ := strconv.Atoi(p.Id)
		if id > li {
			li = id
		}

		if len(p.Path) > lp {
			lp = len(p.Path)
		}
	}
	return ln, li, lp
}
