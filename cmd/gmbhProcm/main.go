package main

import (
	"flag"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/remote"
)

type configFlags []string

func (i *configFlags) String() string {
	return "config path arrray array"
}

func (i *configFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	var configPaths configFlags

	verbose := flag.Bool("verbose", false, "print all output to stdOut and stdErr")
	remoteMode := flag.Bool("remote", false, "start a remote process manager")
	flag.Var(&configPaths, "config", "list to config files")
	flag.Parse()

	// start a remote process manager
	if *remoteMode {
		notify.SetTag("[remote] ")
		rem, _ := remote.NewRemote(defaults.PM_ADDRESS, *verbose)
		for _, path := range configPaths {
			pid, err := rem.AddService(path)
			if err != nil {
				notify.StdMsgErr("could not start service; err=" + err.Error())
			} else {
				notify.StdMsgBlue("service started; pid=" + pid)
			}
		}
		rem.Start()
	} else {

		// start a process manager
		notify.SetTag("[procm] ")
		p := NewProcessManager("", *verbose)
		err := p.Start()
		if err != nil {
			panic(err)
		}
		p.Wait()
	}
}
