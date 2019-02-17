package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// Version of gmbh{Core,Procm,Data}
	Version = "0.9.2"

	// Code name of gmbh release
	Code = "launching"
)

const (
	// Localhost string
	Localhost = "localhost"

	// ServicePort is starting port for assignment to services
	ServicePort = 49502

	// RemotePort is tarting port for assignment to procm remotes
	RemotePort = 59502
)

// DefaultSystemProcm holds the default procm settings
var DefaultSystemProcm = &SystemProcm{
	Address:   "localhost:59500",
	KeepAlive: duration{time.Second * 45},
	Verbose:   true,
	BinPath:   filepath.Join(os.Getenv("$GOPATH"), "bin", "gmbhProcm"),
}

// DefaultSystemCore holds default core settings
var DefaultSystemCore = &SystemCore{
	Mode:      "local",
	Verbose:   true,
	Daemon:    false,
	Address:   "localhost:49500",
	KeepAlive: duration{time.Second * 45},
	BinPath:   filepath.Join(os.Getenv("$GOPATH"), "bin", "gmbhCore"),
}

// DefaultSystemServices holds default services settings
var DefaultSystemServices = &SystemServices{
	Services:          []string{},
	ServicesDirectory: "",
}

// DefaultSystemConfig is the complete default system config
var DefaultSystemConfig = SystemConfig{
	Core:     DefaultSystemCore,
	Procm:    DefaultSystemProcm,
	Services: DefaultSystemServices,
}
