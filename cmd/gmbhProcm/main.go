package main

import (
	"flag"
	"path/filepath"

	"github.com/gmbh-micro/config"
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
		rem, _ := remote.NewRemote(config.DefaultSystemProcm.Address, *verbose)
		for _, path := range configPaths {
			absPath, err := filepath.Abs(path)
			if err != nil {
				notify.LnRedF("could not create abs path to %s", path)
				continue
			}
			rem.AddService(absPath)
		}
		rem.Start()
	} else {

		// start a process manager
		p := NewProcessManager("", *verbose)
		err := p.Start()
		if err != nil {
			panic(err)
		}
		p.Wait()
	}
}
