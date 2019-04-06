package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// Version of gmbh{Core,Procm,Data}
	Version = "0.9.4"

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
	ProcmBinPathWindows = os.Getenv("GOPATH") + "/bin/gmbhProcm.exe"

	// CoreBinPathLinux runtime.GOOS == windows
	CoreBinPathWindows = os.Getenv("GOPATH") + "/bin/gmbhCore.exe"
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

// DefaultSystemConfig is the complete default system config
var DefaultSystemConfig = SystemConfig{
	Core:    DefaultSystemCore,
	Procm:   DefaultSystemProcm,
	Service: make([]*ServiceConfig, 0),
}

///////////////////////////////////////////////////////////////////////////////////
// System Convenience /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////

var (
	// InternalFiles is the relative path from a project config file to where
	// things should be stored such as manifests and logs
	InternalFiles = "gmbh"

	// LogPath is the path from the project directory in which logs should be stored
	LogPath = filepath.Join(InternalFiles, "logs")

	// ManifestPath is the path from the project directory in which manifest toml files
	// should be stored
	ManifestPath = filepath.Join(InternalFiles, "manifest")
)

const (
	// ProcmLogName for log file at Log Path
	ProcmLogName = "procm.log"

	// CoreLogName for log file at Log Path
	CoreLogName = "coreData.log"

	// StdoutExt is the extensions for stdout files
	StdoutExt = "-stdout.log"

	// DefaultServiceLogName ;
	DefaultServiceLogName = "stdout.log"
)
