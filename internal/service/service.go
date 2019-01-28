package service

import (
	"errors"
	"os"
	"strconv"
	"syscall"

	"github.com/gmbh-micro/pmgmt"
)

/*
 * serviceHandler.go
 * Abe Dick
 * Jan 2019
 */

/*
 * Things that still need to be addressed (in no order)
 *
 * Name collisions
 * Ex post facto service registration
 * -> For hotswap, upgrades without restarting CORE
 * Failed process handshake
 * -> Removing failed processes?
 * -> Restarting failed processes?
 * Permissions
 * -> Sending updates
 * -> Mechanism to control updates
 * Querying status of all connected services
 */

// ServiceHandler has methods attached to control adding, removing and
// other service control type things (searching)
type ServiceHandler struct {
	Services map[string]*ServiceControl
	Names    []string
}

// ServiceControl for new method of service discovery.
// The goal is to consolidate the data structure holding both services
// as they are discovered and services as they come online via
// the new service discovery method
type ServiceControl struct {
	Name    string
	Aliases []string
	Config  string

	// ConfigPath is the absolute path to the main directory of the service
	ConfigPath string

	// BinPath is the absolute path to the binary of the service
	BinPath string
	Static  *StaticControl
	Process *pmgmt.Process
	Address string
}

// StaticControl things at service discovery
// (Stored in YAML config file)
type StaticControl struct {
	Name     string   `yaml:"name"`
	Aliases  []string `yaml:"aliases"`
	Language string   `yaml:"language"`
	Makefile bool     `yaml:"makefile"`
	Path     string   `yaml:"pathtobin"`
	IsClient bool     `yaml:"isClient"`
	IsServer bool     `yaml:"isServer"`
}

func NewServiceControl(stat *StaticControl) *ServiceControl {
	return &ServiceControl{
		Name: stat.Name,
		// ConfigPath: stat.Path,
		Aliases: stat.Aliases,
		Static:  stat,
	}
}

func (s *ServiceHandler) AddService(newService *ServiceControl) error {

	if _, ok := s.Services[newService.Name]; ok {
		if ok {
			return errors.New("duplicate service with same name found")
		}
	}

	for _, alias := range newService.Aliases {
		if _, ok := s.Services[alias]; ok {
			if ok {
				return errors.New("duplicate service with same alias found")
			}
		}
	}

	s.Names = append(s.Names, newService.Name)
	s.Services[newService.Name] = newService
	for _, alias := range newService.Aliases {
		s.Services[alias] = newService
	}
	return nil
}

func (s *ServiceHandler) RemoveService() {

}

func (s *ServiceHandler) GetAddress(name string) (string, error) {
	rv := s.Services[name]
	if rv == nil {
		return "", errors.New("could not find service with requested name")
	}
	return rv.Address, nil
}

func (s *ServiceHandler) KillAllServices() {
	for _, n := range s.Names {
		err := raise(s.Services[n].Process.Runtime.Pid, syscall.SIGINT)
		if err == nil {
			// gprint.Ln("Successfully signaled shutdown: "+n, 0)
			// notify.StdMsgMagenta("Successfully signaled shutdown: " + n)
		}
	}
}

func (s *ServiceHandler) GetService(name string) (*ServiceControl, error) {
	service := s.Services[name]
	if service == nil {
		return nil, errors.New("could not find service with name = " + name)
	}
	return service, nil
}

func StartService(service *ServiceControl) (string, error) {

	if service.Static.Language == "go" {
		// fmt.Println(service.ConfigPath)
		service.Process = pmgmt.NewGoProcess(service.Name, service.BinPath, service.ConfigPath)
		pid, err := service.Process.Controller.Start(service.Process)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(pid), nil
	}

	return "", nil
}

func raise(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}
