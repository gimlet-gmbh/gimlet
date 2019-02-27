package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

///////////////////////////////////////////////////////////////////////////////////
// gmbh Project config
///////////////////////////////////////////////////////////////////////////////////

// SystemConfig - config for gmbh
type SystemConfig struct {
	Core        *SystemCore      `toml:"core"`
	Procm       *SystemProcm     `toml:"procm"`
	Service     []*ServiceConfig `toml:"service"`
	Fingerprint string           `toml:"fingerprint"`
}

// SystemCore stores gmbhCore settings
type SystemCore struct {
	Mode      string   `toml:"mode"`
	Verbose   bool     `toml:"verbose"`
	Daemon    bool     `toml:"daemon"`
	Address   string   `toml:"address"`
	KeepAlive duration `toml:"keep_alive"`
	BinPath   string   `toml:"core_bin"`
}

// SystemProcm stores gmbhProcm settings
type SystemProcm struct {
	Address   string   `toml:"address"`
	KeepAlive duration `toml:"keep_alive"`
	Verbose   bool     `toml:"verbose"`
	BinPath   string   `toml:"core_bin"`
}

// ServiceConfig is the static data needed to launch a service from the service launcher
type ServiceConfig struct {
	Args     []string `toml:"args"`
	Env      []string `toml:"env"`
	Language string   `toml:"language"`
	BinPath  string   `toml:"bin_path"`
	SrcPath  string   `toml:"src_path"`
	ProjPath string
}

// ParseSystemConfig parses the entire system config from the file passed in
// otherwise returns an error
func ParseSystemConfig(configFile string) (*SystemConfig, error) {
	var system SystemConfig
	if _, err := toml.DecodeFile(configFile, &system); err != nil {
		return nil, err
	}
	setDefaults(&system)
	return &system, nil
}

// ParseSystemCore returns only the core settings
func ParseSystemCore(configFile string) (*SystemCore, error) {
	system, err := ParseSystemConfig(configFile)
	if err != nil {
		return nil, err
	}
	return system.Core, nil
}

// ParseSystemProcm returns only the procm settings
func ParseSystemProcm(configFile string) (*SystemProcm, error) {
	system, err := ParseSystemConfig(configFile)
	if err != nil {
		return nil, err
	}
	return system.Procm, nil
}

// ParseServices returns only the services and the fingerprint
func ParseServices(configFile string) ([]*ServiceConfig, string, error) {
	system, err := ParseSystemConfig(configFile)
	if err != nil {
		return nil, "", err
	}
	return system.Service, system.Fingerprint, nil
}

// Verify that a service config is balid
func (s *ServiceConfig) Verify() error {
	if s.BinPath == "" && (s.Language == "" || s.SrcPath == "") {
		return fmt.Errorf("must specify a bin_path or language and src_path")
	}
	return nil
}

// setDefaults fills in the blanks with default settings on the important config values
func setDefaults(c *SystemConfig) {
	if c.Core != nil {
		if c.Core.Mode == "" {
			c.Core.Mode = DefaultSystemCore.Mode
		}
		if c.Core.Address == "" {
			c.Core.Address = DefaultSystemCore.Address
		}
		if c.Core.BinPath == "" {
			c.Core.BinPath = DefaultSystemCore.BinPath
		}
	}
	if c.Procm != nil {
		if c.Procm.Address == "" {
			c.Procm.Address = DefaultSystemProcm.Address
		}
		if c.Procm.BinPath == "" {
			c.Procm.BinPath = DefaultSystemProcm.BinPath
		}
	}
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
