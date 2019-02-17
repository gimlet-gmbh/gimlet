package config

import (
	"errors"

	"github.com/BurntSushi/toml"
)

///////////////////////////////////////////////////////////////////////////////////
// gmbh Service config
///////////////////////////////////////////////////////////////////////////////////

// ServiceConfig - for configuring gmbh services
type ServiceConfig struct {
	Mode   string         `toml:"mode"`
	Static *ServiceStatic `toml:"connecting"`
}

// ServiceStatic stores static information needed to start launch a process
type ServiceStatic struct {
	Name     string   `toml:"name"`
	Aliases  []string `toml:"aliases"`
	Args     []string `toml:"args"`
	Env      []string `toml:"env"`
	Language string   `toml:"language"`
	BinPath  string   `toml:"bin_path"`
	SrcPath  string   `toml:"src_path"`
}

// ParseServiceConfig parses the entire service config from the file passed in
// otherwise returns an error
func ParseServiceConfig(path string) (*ServiceConfig, error) {
	var service ServiceConfig
	if _, err := toml.DecodeFile(path, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// ParseServiceStatic returns only the static data from the config
func ParseServiceStatic(path string) (*ServiceStatic, error) {
	g, e := ParseServiceConfig(path)
	if e != nil {
		return nil, e
	}
	return g.Static, nil
}

// Validate returns an error if the config object does not contain enough information to
// be started by the process manager or attached to core
func (s *ServiceStatic) Validate() error {
	if s.Name == "" {
		return errors.New("gmbhCore.NoNameError")
	}
	if s.BinPath != "" {
		return nil
	}
	if s.Language != "" && s.SrcPath != "" {
		return nil
	}
	return errors.New("gmbhProcm.NoServiceInformationError")
}
