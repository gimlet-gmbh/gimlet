package router

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"
)

// Router represents the handling of services including their process
type Router struct {
	Services  map[string]*service.Service
	smLock    *sync.Mutex
	Names     []string
	addresses *addressHandler
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	return &Router{
		Services: make(map[string]*service.Service),
		smLock:   &sync.Mutex{},
		Names:    make([]string, 0),
		addresses: &addressHandler{
			host: defaults.BASE_ADDRESS,
			port: defaults.BASE_PORT,
		},
	}
}

// LookupService looks through the services map and returns the service if it exists
func (r *Router) LookupService(name string) (*service.Service, error) {
	service := r.Services[name]
	if service == nil {
		return nil, errors.New("router.LookupService: could not find service with name  = " + name)
	}
	if service.Process.GetStatus() {
		return service, nil
	}
	return nil, errors.New("router.LookupService: process reported as not running from process management")
}

// LookupAddress looks through the servuce and returns the service address if it could be
// found
func (r *Router) LookupAddress(name string) (string, error) {
	service, err := r.LookupService(name)
	if err != nil {
		return "", err
	}
	if service.Process.GetStatus() {
		return service.Address, nil
	}
	return "", errors.New("router.LookupAddress: process reported as not running from process management")
}

// AddService attaches a service to gmbH
func (r *Router) AddService(configFilePath string) (*service.Service, error) {

	newService, err := service.NewService(configFilePath)
	if err != nil {
		return nil, errors.New("router.AddService.newService " + err.Error())
	}

	// if working with a server, give it an address
	if newService.Static.IsServer {
		newService.Address = r.addresses.assignAddress()
	}

	err = r.addToMap(newService)
	if err != nil {
		return nil, errors.New("router.AddService.addToMap " + err.Error())
	}

	return newService, nil
}

// addToMap returns an error if there is a name or alias conflict with an existing
// service in the service map, otherwise the service's name and alias are added to
// the map
func (r *Router) addToMap(newService *service.Service) error {

	if _, ok := r.Services[newService.Static.Name]; ok {
		return errors.New("duplicate service with same name found")
	}

	for _, alias := range newService.Static.Aliases {
		if _, ok := r.Services[alias]; ok {
			return errors.New("duplicate service with same alias found")
		}
	}

	r.Services[newService.Static.Name] = newService
	r.Names = append(r.Names, newService.Static.Name)
	for _, alias := range newService.Static.Aliases {
		r.Services[alias] = newService
	}

	return nil
}

// KillAllServices sends sigint to all services attached in the service map that
// have a PID
func (r *Router) KillAllServices() {
	for _, name := range r.Names {
		r.raise(r.Services[name].Process.GetRuntime().Pid, syscall.SIGINT)
	}
}

// raise finds a process by pid and then sends sig to it
func (r *Router) raise(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

// TakeInventory returns a list of paths to services
func (r *Router) TakeInventory() []string {
	r.smLock.Lock()
	defer r.smLock.Unlock()

	paths := []string{}
	for _, s := range r.Names {
		paths = append(paths, r.Services[s].Path)
	}
	return paths
}

// addressHandler is in charge of assigning addressses to services
type addressHandler struct {
	table map[string]string
	host  string
	port  int
}

func (a *addressHandler) assignAddress() string {
	addr := a.host + ":" + strconv.Itoa(a.port)
	a.setNextAddress()
	return addr
}

func (a *addressHandler) setNextAddress() {
	a.port += 10
}
