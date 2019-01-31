package service

import (
	"errors"
	"strconv"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
)

// Mode is the mode that the process starts in when registering
type Mode int

const (
	// Ephemeral mode assumes that no enforcement of data types will be made at either
	// end of the gRPC calls between the gimlet-gmbh package while services exchange
	// data.
	//
	// Ephemeral mode works much like http handlers work in the http package of go.
	Ephemeral Mode = 0

	// Custom mode assumes that a shared structure will be used between both services.
	//
	// TODO: Semantics of how to enforce data modes
	Custom Mode = 1
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
		ID:     assignNextID(),
		Path:   dir,
		Static: staticData,
	}
	return &service, nil
}

// StartService attempts to fork/exec service and returns the pid, else error
func (s *Service) StartService() (pid string, err error) {

	if s.Static.Language == "go" {
		s.Process = process.NewGoProc(s.Static.Name, s.createAbsPathToBin(s.Path, s.Static.BinPath), s.Path)
		s.ActiveProcess = true // have to include this because cannot have pointer to interface type in Go
		pid, err := s.Process.Start()
		if err != nil {
			return "-1", errors.New("service.StartService - could not start service: " + err.Error())
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

// assignNextID increments idtag and returns it as a string
func assignNextID() string {
	idtag++
	return strconv.Itoa(idtag)
}
