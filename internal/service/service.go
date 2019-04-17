package service

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gmbh-micro/config"
	"github.com/gmbh-micro/fileutil"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/service/process"
)

// Mode represents how gmbh interacts with the process of the service
type Mode int

const (
	// Remote mode; as in not having been launched from the service launcher
	Remote Mode = 1 + iota

	// Managed mode; as in having been launched from the service launcher
	Managed
)

var modes = [...]string{
	"remote",
	"managed",
}

func (m Mode) String() string {
	if Remote <= m && m <= Managed {
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

	// LogPath is where the std{out,err} of service's process' output will be
	// directed
	LogPath string

	// Static data associated with the service
	// Static *config.ServiceStatic
	Static *config.ServiceConfig

	// If managed, Process will hold the process controller
	Process process.Manager
}

// NewService tries to parse the required info from a config file located at path
func NewService(id string, conf *config.ServiceConfig) (*Service, error) {

	err := conf.Verify()
	if err != nil {
		return nil, err
	}

	service := Service{
		ID:      id,
		Created: time.Now(),
		Path:    filepath.Dir(conf.BinPath),
		Static:  conf,
	}

	if conf.SrcPath != "" {
		service.Path = conf.SrcPath
	}

	return &service, nil
}

// Start attempts to fork/exec service and returns the pid, else error
// service must be in managed or remote mode
func (s *Service) Start(mode string, verbose bool) (pid string, err error) {

	conf := &process.LocalProcessConfig{
		Dir:    s.Path,
		Args:   s.Static.Args,
		Env:    append(os.Environ(), s.Static.Env...),
		Signal: syscall.SIGINT,
	}

	switch s.Static.Language {
	case "node":
		conf.Path = s.Static.SrcPath
		conf.Entry = s.Static.EntryPoint
	// case "go":
	// 	conf.Path = s.Static.SrcPath
	// case "python":

	default:
		conf.Path = s.createAbsPathToBin(s.Path, s.Static.BinPath)
	}

	// in managed mode, a log file is set to capture stdout and stderr
	if mode == "M" {
		s.Mode = Managed
		conf.Signal = syscall.SIGUSR2
		if !verbose {
			if s.Static.ProjPath != "" {
				base := s.Static.BinPath
				if s.Static.SrcPath != "" {
					base = s.Static.SrcPath
				}
				fname := filepath.Base(base) + config.StdoutExt
				s.LogPath = filepath.Join(s.Static.ProjPath, config.LogPath, fname)
			} else {
				s.LogPath = filepath.Join(s.Path, config.LogPath, config.DefaultServiceLogName)
			}
			// notify.LnMagentaF("log at %s", s.LogPath)
			var e error
			conf.LogF, e = fileutil.OpenFile(s.LogPath)
			if e != nil {
				// notify.LnRedF("Error creating log")
			}
		} else {
			// notify.LnMagentaF("verbose; service output directed to os.stdout")
		}
	} else {
		s.Mode = Remote
	}
	switch s.Static.Language {
	case "node":
		interpreter := config.NodeInterpreter
		if mode == "C" {
			interpreter = config.NodeInterpreterAlpine
		}
		s.Process = process.NewInterpretedManager(conf, process.Node, interpreter)
	// case "go":
	// 	s.Process = process.NewInterpretedManager(conf, process.Go)
	default:
		s.Process = process.NewBinaryManager(conf)
	}
	p, err := s.Process.Start()
	if err != nil {
		notify.LnMagentaF("failed to start; err=%s", err.Error())
		return "-1", errors.New("service.StartService.couldNotStartNewService")
	}
	return strconv.Itoa(p), nil

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

// createAbsPathToBin attempts to resolve an absolute path to the binary file to start
func (s *Service) createAbsPathToBin(path, binPath string) string {
	if binPath[0] == '.' {
		return fileutil.GetAbsFpath(binPath)
		// return path + binPath[1:]
	}
	return binPath
}
