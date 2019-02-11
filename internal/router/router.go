package router

import (
	"errors"
	"strconv"
	"sync"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/service"

	"github.com/gmbh-micro/service/process"
	"github.com/gmbh-micro/service/procm"
	"github.com/gmbh-micro/service/static"
)

// Note ARD 2-10
// The difference between services being planetary or managed is the reference to the process manager
// in the service. If there is not one found then planetary and will be rechecked every so
// often

// The maps will be from name -> ( service | procm.Manager )
// The string is the name of the service in both cases

// Router represents the internal handling of services and process managers
type Router struct {
	Services     map[string]*service.Service
	serviceNames []string
	serviceID    int

	processManagers map[string]*procm.Manager
	procmID         int

	addresses *addressHandler
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	return &Router{
		Services:     make(map[string]*service.Service),
		serviceNames: make([]string, 0),
		serviceID:    0,

		processManagers: make(map[string]*procm.Manager),
		procmID:         100,

		addresses: &addressHandler{
			host: defaults.BASE_ADDRESS,
			port: defaults.BASE_PORT,
			mu:   &sync.Mutex{},
		},
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
	for _, name := range r.serviceNames {
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

// AddPlanetaryService to the router
func (r *Router) AddPlanetaryService(staticData *static.Static) (*service.Service, error) {

	newService, err := service.NewPlanetaryService(staticData)
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

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.Services[newService.Static.Name] = newService
	r.serviceNames = append(r.serviceNames, newService.Static.Name)
	for _, alias := range newService.Static.Aliases {
		r.Services[alias] = newService
	}

	return nil
}

// GetAllServices in the service map
func (r *Router) GetAllServices() []*service.Service {
	ret := []*service.Service{}
	for _, s := range r.serviceNames {
		ret = append(ret, r.Services[s])
	}
	return ret
}

// KillAllServices that are currently in managed mode
func (r *Router) KillAllServices() {
	for _, name := range r.serviceNames {
		if r.Services[name].Mode == service.Managed {
			r.Services[name].KillProcess()
		}
	}
}

// RestartAllServices that are currently in managed mode
func (r *Router) RestartAllServices() {
	for _, name := range r.serviceNames {
		if r.Services[name].Mode == service.Managed {
			r.Services[name].RestartProcess()
		} else if r.Services[name].Mode == service.Remote {
			r.processManagers[name].RestartProcess()
		}
	}
}

// GetAllRemoteServices returns a pointer to all services in remote mode
func (r *Router) GetAllRemoteServices() []*service.Service {
	remote := []*service.Service{}
	for _, name := range r.serviceNames {
		if r.Services[name].Mode == service.Remote {
			remote = append(remote, r.Services[name])
		}
	}
	return remote
}

// TakeInventory returns a list of paths to services
func (r *Router) TakeInventory() []string {
	paths := []string{}
	for _, s := range r.serviceNames {
		paths = append(paths, r.Services[s].Path)
	}
	return paths
}

// Reconcile matches services to process managers
func (r *Router) Reconcile() {
	for _, n := range r.serviceNames {
		if r.Services[n].Mode == service.Planetary {
			c := r.processManagers[n]
			if c != nil {
				r.Services[n].Parent = c
				r.Services[n].Mode = service.Remote
			}
		}
	}
}

// AddProcessManager to the router or return it if it already exists
func (r *Router) AddProcessManager(name string) (*procm.Manager, error) {

	exists, _ := r.LookupProcessManager(name)
	if exists != nil {
		return exists, nil
	}

	c := procm.New(name)
	c.ID = r.assignNextProcessManagerID()
	c.Address = r.addresses.assignAddress(false)
	err := r.addProcessManagerToMap(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LookupProcessManager from the router
func (r *Router) LookupProcessManager(name string) (*procm.Manager, error) {
	if r.processManagers[name] == nil {
		return nil, errors.New("router.LookupProcessManager.notFound")
	}
	return r.processManagers[name], nil
}

// LookupProcessManagerByID from the router
func (r *Router) LookupProcessManagerByID(id string) (*procm.Manager, error) {
	for _, c := range r.processManagers {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, errors.New("router.LookupProcessManagerByID.notFound")
}

// GetAllProcessManagers that have been registered to gmbh
func (r *Router) GetAllProcessManagers() []*procm.Manager {
	ret := []*procm.Manager{}
	for _, c := range r.processManagers {
		ret = append(ret, c)
	}
	return ret
}

// addProcessManagerToMap after checking for collisions between names
//
// TODO: Before declaring a duplicate, find a way to actually check before.
//		 This might involve a query to see if the expected answer comes back
//		 from the process manager if we request it...
func (r *Router) addProcessManagerToMap(c *procm.Manager) error {

	if _, ok := r.processManagers[c.Name]; ok {
		return nil
	}

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.processManagers[c.Name] = c

	return nil
}

func (r *Router) assignNextProcessManagerID() string {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.procmID++
	return "c" + strconv.Itoa(r.procmID)
}

func (r *Router) assignNextServiceID() string {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.serviceID++
	return strconv.Itoa(r.serviceID)
}

// addressHandler is in charge of assigning addresses to services
type addressHandler struct {
	table map[string]string
	host  string
	port  int
	mu    *sync.Mutex
}

func (a *addressHandler) assignAddress(service bool) string {
	a.setNextAddress()
	addr := a.host + ":" + strconv.Itoa(a.port)
	return addr
}

func (a *addressHandler) setNextAddress() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.port += 2
}
