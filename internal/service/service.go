package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
)

// Mode represents how gmbh interacts with the process of the service
type Mode int

const (
	// Managed mode is for services whose underlying process is managed by gmbhCore
	Managed Mode = 1 + iota

	// Remote mode is for services whose underlying process is mangaged by gmbhContainer
	Remote

	// Planetary mode is for services whose underlying process is not mangaged by any gmbh tooling
	Planetary

	// GmbH mode is for managing the gmbh process itself
	GmbH
)

var modes = [...]string{
	"Managed",
	"Remote",
	"Planetary",
	"GmbH",
}

func (m Mode) String() string {
	if Managed <= m && m <= GmbH {
		return modes[m-1]
	}
	return "%!Mode()"
}

// Service represents a service including all static and runtime data
type Service struct {
	// The ephemeral id of the service. Note that ID is mutable and changes when a
	// planetary service becomes a remote service
	ID      string
	Path    string
	Created time.Time
	Address string
	Mode    Mode
	Logs    *notify.Log

	// Static data associated with the service
	Static *static.Static

	// If managed, Process will hold the process controller
	Process process.Manager

	// If Remote, Remote will hold the remote process controller
	Remote *process.RemoteManager
}

// NewService tries to parse the required info from a config file located at path
func NewService(id, path string) (*Service, error) {
	staticData, err := static.ParseData(path)
	if err != nil {
		return nil, err
	}

	dir := path[:len(path)-len(defaults.CONFIG_FILE)]

	service := Service{
		ID:      id,
		Created: time.Now(),
		Mode:    Managed,
		Path:    dir,
		Static:  staticData,
	}

	ok := static.DataIsValid(staticData)
	if !ok {
		return nil, errors.New("invalid config file")
	}
	return &service, nil
}

// NewManagedService tries to parse the required info from a config file located at path
func NewManagedService(id, path string) (*Service, error) {
	staticData, err := static.ParseData(path)
	if err != nil {
		return nil, err
	}

	dir := path[:len(path)-len(defaults.CONFIG_FILE)]

	service := Service{
		ID:      id,
		Created: time.Now(),
		Mode:    Managed,
		Path:    dir,
		Static:  staticData,
	}

	ok := static.DataIsValid(staticData)
	if !ok {
		return nil, errors.New("invalid config file")
	}
	return &service, nil
}

// NewPlanetaryService returns a new service with static data that is passed in
func NewPlanetaryService(id string, staticData *static.Static) (*Service, error) {
	if staticData == nil {
		return nil, errors.New("static data not present")
	}
	service := Service{
		ID:      id,
		Created: time.Now(),
		Mode:    Planetary,
		Static:  staticData,
	}
	return &service, nil
}

// Start attempts to fork/exec service and returns the pid, else error
// service must be in managed or remote mode
func (s *Service) Start() (pid string, err error) {

	if s.Mode == Planetary || s.Mode == Remote {
		return "-1", errors.New("service.StartService.invalidServiceMode")
	}
	if s.Static.Language == "go" {

		s.Process = process.NewLocalBinaryManager(s.Static.Name, s.createAbsPathToBin(s.Path, s.Static.BinPath), s.Path, []string{}, []string{})
		pid, err := s.Process.Start()
		if err != nil {
			notify.StdMsgDebug("failed to start")
			return "-1", errors.New("service.StartService.couldNotStartNewService")
		}
		return strconv.Itoa(pid), nil

	} else if s.Static.Language == "node" {
		return "-1", errors.New("service.StartService.nodeNotYetSupported")
	} else if s.Static.Language == "python" {
		return "-1", errors.New("service.StartService.pythonNotYetSupported")
	}

	return "-1", errors.New("service.StartService.invalidLanguage")
}

// Restart if it is in mangaed or attached mode
func (s *Service) Restart() (string, error) {
	if s.Mode == Planetary {
		return "-1", errors.New("Service.RestartProcess.inPlanetaryMode")
	}

	if s.Mode == Remote {
		return s.Remote.RestartProcess()
	}

	// s.Mode == Managed
	pid, err := s.Process.Restart(false)
	if err != nil {
		return "-1", err
	}
	return strconv.Itoa(pid), nil
}

// Kill a process
func (s *Service) Kill() {
	if s.Mode == Managed {
		s.Process.Kill(true)
	}
}

// StartLog starts the logger for process management information
func (s *Service) StartLog(path, filename string) {
	s.Logs = notify.NewLogFile(path, filename, false)
}

// createAbsPathToBin attempts to resolve an absolute path to the binary file to start
func (s *Service) createAbsPathToBin(path, binPath string) string {
	absPath := ""
	if binPath[0] == '.' {
		absPath = path + binPath[1:]
	}
	return absPath
}

// Println adds a log message to the service's log if it has been configured
func (s *Service) Println(msg string) {
	if s.Logs != nil {
		s.Logs.Ln(msg)
	}
}

// GetMode returns the mode as a string
func (s *Service) GetMode() string {
	if s.Mode == Managed {
		return "managed"
	}
	if s.Mode == Remote {
		return "remote"
	}

	if s.Mode == Planetary {
		return "planetary"
	}

	return ""
}
