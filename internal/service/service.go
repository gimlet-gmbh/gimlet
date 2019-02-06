package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/log"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
)

// Mode represents how gmbh interacts with the process of the service
type Mode int

const (
	// Managed mode is for services whose underlying process is managed by gmbhCore
	Managed Mode = 2

	// Remote mode is for services whose underlying process is mangaged by gmbhContainer
	Remote Mode = 3

	// Planetary mode is for services whose underlying process is not mangaged by any gmbh tooling
	Planetary Mode = 4

	// GmbH mode is for managing the gmbh process itself
	GmbH Mode = 5
)

// Status is the enumerated stated of the current status of the service, !!not the process!!
type Status int

const (
	// Configured ; the service's config has been parsed and is valid
	Configured Status = 1

	// Unconfigured ; the config file was not able to be parsed
	Unconfigured Status = 2
)

var idtag int

func init() {
	idtag = defaults.STARTING_ID
}

// Service represents a service including all static and runtime data
type Service struct {
	ID            string
	Path          string
	Address       string
	Mode          Mode
	Status        Status
	Static        *static.Static
	Process       process.Process
	ActiveProcess bool
	Logs          *log.Log
}

// GetProcess returns the process or empty process and error of a service
func (s *Service) GetProcess() process.Process {
	if s.Mode == Planetary {
		return process.NewEmptyProc()
	}
	if !s.ActiveProcess {
		return process.NewEmptyProc()
	}
	return s.Process
}

// NewManagedService tries to parse the required info from a config file located at path
func NewManagedService(path string) (*Service, error) {
	staticData, err := static.ParseData(path)
	if err != nil {
		return nil, err
	}

	dir := path[:len(path)-len(defaults.CONFIG_FILE)]

	service := Service{
		ID:     assignNextID(),
		Mode:   Managed,
		Path:   dir,
		Static: staticData,
	}

	ok := static.DataIsValid(staticData)
	if !ok {
		service.Status = Unconfigured
		return &service, errors.New("invalid config file")
	}
	service.Status = Configured
	return &service, nil
}

// NewRemoteService returns a new service with static data that is passed in
func NewRemoteService(staticData *static.Static) (*Service, error) {

	service := Service{
		ID:     assignNextID(),
		Mode:   Remote,
		Static: staticData,
	}

	service.Status = Configured
	return &service, nil
}

// StartService attempts to fork/exec service and returns the pid, else error
// service must be in managed or remote mode
func (s *Service) StartService() (pid string, err error) {

	if s.Mode == Planetary {
		return "-1", errors.New("service.StartService.invalidService Mode")
	}

	if s.Static.Language == "go" {
		s.Process = process.NewGoProc(s.Static.Name, s.createAbsPathToBin(s.Path, s.Static.BinPath), s.Path)
		s.ActiveProcess = true // have to include this because cannot have pointer to interface type in Go
		pid, err := s.Process.Start()
		if err != nil {
			s.LogMsg("error starting service; err=" + err.Error())
			return "-1", errors.New("service.StartService.couldNotStartService")
		}
		s.LogMsg("started with pid=" + strconv.Itoa(pid))
		return strconv.Itoa(pid), nil
	} else if s.Static.Language == "node" {
		s.Process = process.NewNodeProc()
		return "-1", errors.New("service.StartService.nodeNotYetSupported")
	} else if s.Static.Language == "python" {
		s.Process = process.NewPyProc()
		return "-1", errors.New("service.StartService.pythonNotYetSupported")
	}

	return "-1", errors.New("service.StartService.invalidLanguage")
}

// KillProcess if it is in managed or attached mode
func (s *Service) KillProcess() {
	if s.Mode == Planetary {
		return
	}
	s.LogMsg("kill process request at " + time.Now().Format(time.RFC3339))
	s.Process.Kill(true)
}

// RestartProcess if it is in mangaed or attached mode
func (s *Service) RestartProcess() {
	if s.Mode == Planetary {
		return
	}
	s.LogMsg("kill process request at " + time.Now().Format(time.RFC3339))
	pid, err := s.Process.Restart(false)
	if err != nil {
		s.LogMsg("error restarting; err=" + err.Error())
	}
	s.LogMsg("restarted with pid=" + strconv.Itoa(pid))
}

// StartLog starts the logger for process management information
func (s *Service) StartLog(path, filename string) {
	s.Logs = log.NewLogFile(path, filename)
}

// createAbsPathToBin attempts to resolve an absolute path to the binary file to start
func (s *Service) createAbsPathToBin(path, binPath string) string {
	absPath := ""
	if binPath[0] == '.' {
		absPath = path + binPath[1:]
	}
	return absPath
}

// assignNextID increments idtag and returns it as a string
func assignNextID() string {
	idtag++
	return strconv.Itoa(idtag)
}

// LogMsg adds a log message to the service's log if it has been configured
func (s *Service) LogMsg(msg string) {
	if s.Logs != nil {
		s.Logs.Msg(msg)
	}
}
