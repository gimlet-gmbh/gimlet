package router

import (
	"errors"
	"strconv"
	"sync"

	"github.com/gmbh-micro/defaults"
	"github.com/gmbh-micro/notify"
	"github.com/gmbh-micro/service"

	"github.com/gmbh-micro/service/process"
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
	remoteManagers map[string]*process.RemoteManager
	services       map[string]*service.Service
	serviceNames   []string
	idCounter      int
	addresses      *addressHandler
	Verbose        bool
}

// NewRouter initializes and returns a new Router struct
func NewRouter() *Router {
	return &Router{
		remoteManagers: make(map[string]*process.RemoteManager),
		services:       make(map[string]*service.Service),
		serviceNames:   make([]string, 0),
		idCounter:      100,
		addresses: &addressHandler{
			host: defaults.BASE_ADDRESS,
			port: defaults.BASE_PORT,
			mu:   &sync.Mutex{},
		},
		Verbose: true,
	}
}

// LookupService looks through the services map and returns the service if it exists
func (r *Router) LookupService(name string) (*service.Service, error) {
	r.verbose("looking up " + name)
	retrievedService := r.services[name]
	if retrievedService == nil {
		r.verbose("not found")
		return nil, errors.New("router.LookupService.NotFound")
	}
	r.verbose("found")
	return retrievedService, nil
}

// LookupServiceAddress looks through the service map and returns the service address if it could be found
func (r *Router) LookupServiceAddress(name string) (string, error) {
	r.verbose("looking up " + name)
	retrievedService := r.services[name]
	if retrievedService == nil {
		r.verbose("not found")
		return "", errors.New("router.LookupService.NotFound")
	}
	r.verbose("found")
	return retrievedService.Address, nil
}

// LookupServiceID looks through the service map and returns the service if it could be matched to id
func (r *Router) LookupServiceID(id string) (*service.Service, error) {
	r.Reconcile()
	r.verbose("looking up " + id)
	for _, name := range r.serviceNames {
		if r.services[name].ID == id {
			r.verbose("found")
			return r.services[name], nil
		}
	}
	r.verbose("not found")
	return nil, errors.New("router.LookupServiceID.notFound")
}

// AddManagedService attaches a service to gmbH
func (r *Router) AddManagedService(configFilePath string) (*service.Service, error) {
	r.verbose("adding new managed service")
	newService, err := service.NewManagedService(r.assignNextID(), configFilePath)
	if err != nil {
		r.verbose("error=" + err.Error())
		return nil, errors.New("router.AddService.newService " + err.Error())
	}
	r.verbose("added " + newService.Static.Name)

	newService.Address = r.addresses.assignAddress(true)
	r.verbose("address=" + newService.Address)

	err = r.addToMap(newService)
	if err != nil {
		return nil, err
	}

	return newService, nil
}

// AddPlanetaryService to the router
func (r *Router) AddPlanetaryService(staticData *static.Static) (*service.Service, error) {
	r.verbose("adding planetary service")
	newService, err := service.NewPlanetaryService(r.assignNextID(), staticData)
	if err != nil {
		r.verbose("error=" + err.Error())
		return nil, err
	}
	r.verbose("added " + newService.Static.Name)

	newService.Address = r.addresses.assignAddress(true)
	r.verbose("address=" + newService.Address)

	err = r.addToMap(newService)
	if err != nil {
		return nil, err
	}
	return newService, nil
}

// addToMap returns an error if there is a name or alias conflict with an existing
// service in the service map, otherwise the service's name and alias are added to
// the map
func (r *Router) addToMap(newService *service.Service) error {

	if _, ok := r.services[newService.Static.Name]; ok {
		r.verbose("could not add to map, duplicate name")
		return errors.New("router.addToMap: duplicate service with same name found")
	}

	for _, alias := range newService.Static.Aliases {
		if _, ok := r.services[alias]; ok {
			r.verbose("could not add to map, duplicate alias=" + alias)
			return errors.New("router.addToMap: duplicate service with same alias found")
		}
	}

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.services[newService.Static.Name] = newService
	r.serviceNames = append(r.serviceNames, newService.Static.Name)
	for _, alias := range newService.Static.Aliases {
		r.services[alias] = newService
	}

	r.verbose("added to map")

	return nil
}

// GetAllServices in the service map
func (r *Router) GetAllServices() []*service.Service {
	r.verbose("get all services")
	ret := []*service.Service{}
	for _, s := range r.serviceNames {
		ret = append(ret, r.services[s])
	}
	r.verbose("retrieved " + strconv.Itoa(len(ret)) + " services")
	return ret
}

// RestartAllServices that are managed or remote
func (r *Router) RestartAllServices() {
	r.Reconcile()
	r.verbose("restarting all services")
	for _, name := range r.serviceNames {
		if r.services[name].Mode == service.Managed || r.services[name].Mode == service.Remote {
			r.verbose("sending restart singal to " + name + " at " + r.services[name].Address)
			r.services[name].Restart()
		}
	}
}

// KillManagedServices that are managed or remote
func (r *Router) KillManagedServices() {
	r.verbose("killing all managed services")
	for _, name := range r.serviceNames {
		if r.services[name].Mode == service.Managed {
			r.verbose("sending shutdown to " + name)
			r.services[name].Kill()
		}
	}
}

// GetAllRemoteServices returns a pointer to all services in remote mode
func (r *Router) GetAllRemoteServices() []*service.Service {
	r.Reconcile()
	remote := []*service.Service{}
	for _, name := range r.serviceNames {

		if r.services[name].Mode == service.Remote {
			r.verbose("checking: " + name + "; mode=remote")
		}
		if r.services[name].Mode == service.Managed {
			r.verbose("checking: " + name + "; mode=managed")
		}
		if r.services[name].Mode == service.Planetary {
			r.verbose("checking: " + name + "; mode=planetary")
		}
		if r.services[name].Mode == service.Remote {
			remote = append(remote, r.services[name])
		}
	}
	return remote
}

// // TakeInventory returns a list of paths to services
// func (r *Router) TakeInventory() []string {
// 	paths := []string{}
// 	for _, s := range r.serviceNames {
// 		paths = append(paths, r.Services[s].Path)
// 	}
// 	return paths
// }

// Reconcile matches services to process managers
func (r *Router) Reconcile() {
	for _, n := range r.serviceNames {
		if r.services[n].Mode == service.Planetary {
			r.verbose("attempting to upgrade " + r.services[n].Static.Name)
			c := r.remoteManagers[n]
			if c != nil {
				r.verbose("found remote manager match")
				r.verbose("upgrading id to " + c.ID + "-" + r.services[n].ID)
				r.services[n].ID = c.ID + "-" + r.services[n].ID
				r.services[n].Remote = c
				r.services[n].Mode = service.Remote
			} else {
				r.verbose("did not find remote manager")
			}
		}
	}
}

// AddRemoteProcessManager to the router or return it if it already exists
func (r *Router) AddRemoteProcessManager(name string) (*process.RemoteManager, error) {

	exists, _ := r.LookupProcessManager(name)
	if exists != nil {
		return exists, nil
	}

	r.verbose("adding process manager for : " + name)

	c := process.NewRemoteManager(name, "c"+r.assignNextID())
	c.Address = r.addresses.assignAddress(false)
	err := r.addProcessManagerToMap(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LookupProcessManager from the router
func (r *Router) LookupProcessManager(name string) (*process.RemoteManager, error) {
	if r.remoteManagers[name] == nil {
		return nil, errors.New("router.LookupProcessManager.notFound")
	}
	return r.remoteManagers[name], nil
}

// LookupProcessManagerID from the router
func (r *Router) LookupProcessManagerID(id string) (*process.RemoteManager, error) {
	for _, c := range r.remoteManagers {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, errors.New("router.LookupProcessManagerByID.notFound")
}

// GetAllProcessManagers that have been registered to gmbh
func (r *Router) GetAllProcessManagers() []*process.RemoteManager {
	ret := []*process.RemoteManager{}
	for _, c := range r.remoteManagers {
		ret = append(ret, c)
	}
	return ret
}

func (r *Router) verbose(msg string) {
	if r.Verbose {
		notify.StdMsgBlueNoPrompt(" [rtr] " + msg)
	}
}

// addProcessManagerToMap after checking for collisions between names
//
// TODO: Before declaring a duplicate, find a way to actually check before.
//		 This might involve a query to see if the expected answer comes back
//		 from the process manager if we request it...
func (r *Router) addProcessManagerToMap(c *process.RemoteManager) error {
	r.verbose("adding process manager to map")
	if _, ok := r.remoteManagers[c.Name]; ok {
		r.verbose("already in map, skip")
		return nil
	}

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	r.verbose("added : " + c.Name)
	r.remoteManagers[c.Name] = c

	return nil
}

func (r *Router) assignNextID() string {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	r.idCounter++
	return strconv.Itoa(r.idCounter)
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
