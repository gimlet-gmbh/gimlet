package service

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/service/process"
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
	ID      string
	Path    string
	Created time.Time
	Address string
	Mode    Mode
	Logs    *notify.Log

	// Static data associated with the service
	Static *config.ServiceStatic

	// If managed, Process will hold the process controller
	Process process.Manager

	// If Remote, Remote will hold the remote process controller
	Remote *process.RemoteManager
}

// NewService tries to parse the required info from a config file located at path
func NewService(id, path string) (*Service, error) {

	staticData, err := config.ParseServiceStatic(path)
	if err != nil {
		return nil, err
	}
	valid := staticData.Validate()
	if valid != nil {
		return nil, err
	}

	service := Service{
		ID:      id,
		Created: time.Now(),
		Mode:    Managed,
		Path:    filepath.Dir(path),
		Static:  staticData,
	}

	return &service, nil
}

// Start attempts to fork/exec service and returns the pid, else error
// service must be in managed or remote mode
func (s *Service) Start(mode string) (pid string, err error) {

	s.Static.Env = append(s.Static.Env, os.Environ()...)

	if s.Static.Language == "go" {
		ssignal := syscall.SIGINT

		if mode == "managed" {
			notify.LnYellowF("using sigusr2 as shutdown signal")
			ssignal = syscall.SIGUSR2
		} else {
			notify.LnYellowF("using sigint as shutdown signal")
		}

		s.Process = process.NewLocalBinaryManager(s.Static.Name, s.createAbsPathToBin(s.Path, s.Static.BinPath), s.Path, s.Static.Args, s.Static.Env, ssignal)
		pid, err := s.Process.Start()
		if err != nil {
			notify.LnYellowF("failed to start; err=%s", err.Error())
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

// Restart the process
func (s *Service) Restart() (string, error) {
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

// EnableGracefulShutdown tells the process to stop attempting to restart
func (s *Service) EnableGracefulShutdown() {
	s.Process.GracefulShutdown()
}

// StartLog starts the logger for process management information
func (s *Service) StartLog(path, filename string) {
	s.Logs = notify.NewLogFile(path, filename, false)
}

// createAbsPathToBin attempts to resolve an absolute path to the binary file to start
func (s *Service) createAbsPathToBin(path, binPath string) string {
	if binPath[0] == '.' {
		return path + binPath[1:]
	}
	return binPath
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
