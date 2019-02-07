package router

import (
	"errors"
	"strconv"
	"sync"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"
	"github.com/gmbh-micro/service/container"
	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/static"
)

// Router represents the handling of services including their process
type Router struct {
	Services      map[string]*service.Service
	serviceLock   *sync.Mutex
	containers    map[string]*container.Container
	containerLock *sync.Mutex
	smLock        *sync.Mutex
	Names         []string
	CNames        []string
	addresses     *addressHandler
	containerID   int
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	return &Router{
		Services:      make(map[string]*service.Service),
		serviceLock:   &sync.Mutex{},
		containers:    make(map[string]*container.Container),
		containerLock: &sync.Mutex{},
		smLock:        &sync.Mutex{},
		Names:         make([]string, 0),
		addresses: &addressHandler{
			host: defaults.BASE_ADDRESS,
			port: defaults.BASE_PORT,
		},
		containerID: 100,
	}
}

// LookupService looks through the services map and returns the service if it exists
func (r *Router) LookupService(name string) (*service.Service, error) {
	retrievedService := r.Services[name]
	if retrievedService == nil {
		return nil, errors.New("router.LookupService.nameNotFound")
	}
	if retrievedService.Mode == service.Managed {
		if retrievedService.GetProcess().GetStatus() == process.Running || retrievedService.GetProcess().GetStatus() == process.Stable {
			return retrievedService, nil
		}
		return retrievedService, errors.New("router.LookupService.processNotRunning")
	}
	return retrievedService, nil
}

// LookupAddress looks through the service map and returns the service address if it could be found
func (r *Router) LookupAddress(name string) (string, error) {
	retrievedService := r.Services[name]
	if retrievedService == nil {
		return "", errors.New("router.LookupService.nameNotFound")
	}
	if retrievedService.Mode == service.Managed {
		if retrievedService.GetProcess().GetStatus() == process.Running || retrievedService.GetProcess().GetStatus() == process.Stable {
			return retrievedService.Address, nil
		}
		return "", errors.New("router.LookupService.processNotRunning")
	}
	return retrievedService.Address, nil
}

// LookupByID looks through the service map and returns the service if the id matches the parameter
func (r *Router) LookupByID(id string) (*service.Service, error) {
	for _, name := range r.Names {
		if r.Services[name].ID == id {
			return r.Services[name], nil
		}
	}
	return nil, errors.New("router.LookupByID: could not find service")
}

// AddManagedService attaches a service to gmbH
func (r *Router) AddManagedService(configFilePath string) (*service.Service, error) {

	newService, err := service.NewManagedService(configFilePath)
	if err != nil {
		return nil, errors.New("router.AddService.newService " + err.Error())
	}

	// if working with a server, give it an address
	if newService.Static.IsServer {
		newService.Address = r.addresses.assignAddress(true)
	}

	err = r.addToMap(newService)
	if err != nil {
		return nil, errors.New("router.AddService.addToMap " + err.Error())
	}

	return newService, nil
}

// AddRemoteService to the router
func (r *Router) AddRemoteService(staticData *static.Static) (*service.Service, error) {

	newService, err := service.NewRemoteService(staticData)
	if err != nil {
		return nil, errors.New("router.AddService.newService " + err.Error())
	}

	// if working with a server, give it an address
	if newService.Static.IsServer {
		newService.Address = r.addresses.assignAddress(true)
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
		return errors.New("router.addToMap: duplicate service with same name found")
	}

	for _, alias := range newService.Static.Aliases {
		if _, ok := r.Services[alias]; ok {
			return errors.New("router.addToMap: duplicate service with same alias found")
		}
	}

	r.Services[newService.Static.Name] = newService
	r.Names = append(r.Names, newService.Static.Name)
	for _, alias := range newService.Static.Aliases {
		r.Services[alias] = newService
	}

	return nil
}

// GetAllServices in the service map
func (r *Router) GetAllServices() []*service.Service {
	ret := []*service.Service{}
	for _, s := range r.Names {
		ret = append(ret, r.Services[s])
	}
	return ret
}

// KillAllServices that are currently in managed mode
func (r *Router) KillAllServices() {
	for _, name := range r.Names {
		if r.Services[name].Mode == service.Managed {
			r.Services[name].KillProcess()
		}
	}
}

// RestartAllServices that are currently in managed mode
func (r *Router) RestartAllServices() {
	for _, name := range r.Names {
		if r.Services[name].Mode == service.Managed {
			r.Services[name].RestartProcess()
		}
	}
}

// GetAllRemoteServices returns a pointer to all services in remote mode
func (r *Router) GetAllRemoteServices() []*service.Service {
	remote := []*service.Service{}
	for _, name := range r.Names {
		if r.Services[name].Mode == service.Remote {
			remote = append(remote, r.Services[name])
		}
	}
	return remote
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

// AddContainer to the router or return the container if it already exists
func (r *Router) AddContainer(name string) (*container.Container, error) {

	exists, _ := r.LookupContainer(name)
	if exists != nil {
		return exists, nil
	}

	c := container.New(name)
	c.ID = r.assignNextContainerID()
	c.Address = r.addresses.assignAddress(false)
	err := r.addContainerToMap(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LookupContainer from the router
func (r *Router) LookupContainer(name string) (*container.Container, error) {
	if r.containers[name] == nil {
		return nil, errors.New("router.LookupContainer.notFound")
	}
	return r.containers[name], nil
}

// addContainerToMap after checking for collisions between names
//
// TODO: Before declaring a duplicate, find a way to actually check before.
//		 This might involve a query to see if the expected answer comes back
//		 from the container if we request it...
func (r *Router) addContainerToMap(c *container.Container) error {
	r.containerLock.Lock()
	defer r.containerLock.Unlock()
	if _, ok := r.containers[c.Name]; ok {
		return nil
		// return errors.New("router.addContainerToMap.duplicate")
	}
	r.containers[c.Name] = c
	r.CNames = append(r.CNames, c.Name)
	return nil
}

func (r *Router) assignNextContainerID() string {
	r.containerID++
	return "c" + strconv.Itoa(r.containerID)
}

// addressHandler is in charge of assigning addresses to services
type addressHandler struct {
	table map[string]string
	host  string
	port  int
}

func (a *addressHandler) assignAddress(service bool) string {
	addr := a.host + ":" + strconv.Itoa(a.port)
	a.setNextAddress()
	return addr
}

func (a *addressHandler) setNextAddress() {
	a.port += 10
}
