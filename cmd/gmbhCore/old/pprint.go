package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/gmbh-micro/cabal"
)

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
		if p.Status != "running" {
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
