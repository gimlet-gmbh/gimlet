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

var (
	// ProcmBinPathMac runtime.GOOS ==  darwin
	ProcmBinPathMac = os.Getenv("GOPATH") + "/bin/gmbhProcm"

	// CoreBinPathMac runtime.GOOS == darwin
	CoreBinPathMac = os.Getenv("GOPATH") + "/bin/gmbhCore"

	// ProcmBinPathLinux runtime.GOOS ==  linux
	ProcmBinPathLinux = os.Getenv("GOPATH") + "/bin/gmbhProcm"

	// CoreBinPathLinux runtime.GOOS == linux
	CoreBinPathLinux = os.Getenv("GOPATH") + "/bin/gmbhCore"

	// ProcmBinPathLinux runtime.GOOS ==  windows
	ProcmBinPathWindows = os.Getenv("GOPATH") + "/bin/gmbhProcm"

	// CoreBinPathLinux runtime.GOOS == windows
	CoreBinPathWindows = os.Getenv("GOPATH") + "/bin/gmbhCore"
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

// DefaultServiceConfig is the default configuration for a service
var DefaultServiceConfig = &ServiceConfig{
	CoreAddress: DefaultSystemCore.Address,
}
