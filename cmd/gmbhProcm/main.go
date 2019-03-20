package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gmbh-micro/config"
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

	coreAddr := config.DefaultSystemProcm.Address
	try := os.Getenv("DOCKERNETWORK")
	if try != "" {
		coreAddr = try
	}

	// start a remote process manager
	if *remoteMode {

		rem, _ := remote.NewRemote(coreAddr, *verbose)
		for _, path := range configPaths {

			sconfs, fingerprint, err := config.ParseServices(path)
			if err != nil {
				panic(err)
			}

			if fingerprint != os.Getenv("FINGERPRINT") {
				panic(fmt.Errorf("fingerprints do not match (%s != %s)", fingerprint, os.Getenv("FINGERPRINT")))
			}

			for _, sconf := range sconfs {
				if os.Getenv("PROJPATH") != "" {
					sconf.ProjPath = os.Getenv("PROJPATH")
				}
				rem.AddService(sconf)
			}
		}
		rem.Start()
	} else {

		// start a process manager
		p := NewProcessManager(coreAddr, *verbose)
		err := p.Start()
		if err != nil {
			panic(err)
		}
		p.Wait()
	}
}
