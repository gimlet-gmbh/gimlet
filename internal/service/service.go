package service

import (
	"errors"
	"strconv"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
	"github.com/rs/xid"
)

// Service represents a service including all static and runtime data
type Service struct {
	ID      string
	Path    string
	Address string
	Static  *static.Static
	Process *process.Process
}

// NewService tries to parse the required info from a config file located at path
func NewService(path string) (*Service, error) {
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
		ID:     xid.New().String(),
		Path:   dir,
		Static: staticData,
	}
	return &service, nil
}

// StartService attempts to fork/exec service and returns the pid, else error
func (s *Service) StartService() (pid string, err error) {

	if s.Static.Language == "go" {

		// fmt.Println(s.createAbsPathToBin(s.Path, s.Static.BinPath))
		s.Process = process.NewGoProcess(s.createAbsPathToBin(s.Path, s.Static.BinPath), "")
		pid, err := s.Process.Control.Start(s.Process)
		if err != nil {
			return "-1", errors.New("could not start service")
		}
		return strconv.Itoa(pid), nil
	}

	return "-1", errors.New("service.StartService not implemented for languages other than go")
}

// createAbsPathToBin attempts to resolve an absolute path to the binary file to start
func (s *Service) createAbsPathToBin(path, binPath string) string {
	absPath := ""
	if binPath[0] == '.' {
		absPath = path + binPath[1:]
	}
	return absPath
}
