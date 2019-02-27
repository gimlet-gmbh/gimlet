package gmbh

import "github.com/gmbh-micro/config"

// Option functions set options from the client
type Option func(*options)

// options contain the runtime configurable parameters
type options struct {

	// RuntimeOptions are options that can be determined at runtime
	runtime *RuntimeOptions

	// standalone options are those intended for use without the service launcher or remotes
	standalone *StandaloneOptions

	// service options are those that are used for identifying the service with core
	service *ServiceOptions
}

// RuntimeOptions - user configurable
type RuntimeOptions struct {
	// Should the client block the main thread until shutdown signal is received?
	Blocking bool

	// Should the client run in verbose mode. in Verbose mode, debug information regarding
	// the gmbh client will be printed to stdOut
	Verbose bool
}

// StandaloneOptions - user configurable, for use only without the service launcher or remotes
type StandaloneOptions struct {
	// The address back to core
	CoreAddress string
}

// ServiceOptions - user configurable, a name must be set, this is how other services will contact this one.
type ServiceOptions struct {
	// Name - the unique name of the service
	Name string

	// Aliases - like the name, must be unique across all services; act as shortcut names
	Aliases []string
}

var defaultOptions = options{
	runtime: &RuntimeOptions{
		Blocking: false,
		Verbose:  false,
	},
	standalone: &StandaloneOptions{
		CoreAddress: config.DefaultSystemCore.Address,
	},
	service: &ServiceOptions{
		Name:    "",
		Aliases: make([]string, 0),
	},
}

// SetRuntime options of the client
func SetRuntime(r RuntimeOptions) Option {
	return func(o *options) {
		o.runtime.Blocking = r.Blocking
		o.runtime.Verbose = r.Verbose
	}
}

// SetStandalone options of the client
func SetStandalone(s StandaloneOptions) Option {
	return func(o *options) {
		o.standalone.CoreAddress = s.CoreAddress
	}
}

// SetService options of the client
func SetService(s ServiceOptions) Option {
	return func(o *options) {
		o.service.Name = s.Name
		o.service.Aliases = s.Aliases
	}
}
