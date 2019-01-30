package service

import (
	"errors"

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
	service := Service{
		ID:     xid.New().String(),
		Path:   path,
		Static: staticData,
	}
	return &service, nil
}

// StartService attempts to fork/exec service and returns the pid, else error
func (s *Service) StartService() (pid string, err error) {

	return "-1", nil
}
