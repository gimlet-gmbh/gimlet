package service

import (
	"errors"
	"strconv"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
)

// Mode represents how gmbh interacts with the process of the service
type Mode int

const (
	// Managed mode is for services whose underlying process is managed by gmbh
	Managed Mode = 2
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
	Static        *static.Static
	Process       process.Process
	ActiveProcess bool
}

// GetProcess returns the process or empty process and error of a service
func (s *Service) GetProcess() process.Process {
	if s.Mode != Managed {
		return process.NewEmptyProc()
		// return process.NewEmptyProc(), errors.New("service.getProcess.unmanagedServiceProcessRequest")
	}
	if !s.ActiveProcess {
		return process.NewEmptyProc()
		// return process.NewEmptyProc(), errors.New("service.getProcess.inactiveProcess")
	}
	return s.Process
	// return s.Process, nil
}

// NewService tries to parse the required info from a config file located at path
func NewService(path string, mode Mode) (*Service, error) {
	staticData, err := static.ParseData(path)
	if err != nil {
		return nil, err
	}
	ok := static.DataIsValid(staticData)
	if !ok {
		return nil, errors.New("invalid config file")
	}

	dir := path[:len(path)-len(defaults.CONFIG_FILE)]

	service := Service{
		ID:     assignNextID(),
		Mode:   mode,
		Path:   dir,
		Static: staticData,
	}
	return &service, nil
}

// StartService attempts to fork/exec service and returns the pid, else error
// service must be in managed mode
func (s *Service) StartService() (pid string, err error) {

	if s.Mode != Managed {
		return "-1", errors.New("service.StartService.invalidService Mode")
	}

	if s.Static.Language == "go" {
		s.Process = process.NewGoProc(s.Static.Name, s.createAbsPathToBin(s.Path, s.Static.BinPath), s.Path)
		s.ActiveProcess = true // have to include this because cannot have pointer to interface type in Go
		pid, err := s.Process.Start()
		if err != nil {
			return "-1", errors.New("service.StartService.couldNotStartService")
		}
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
