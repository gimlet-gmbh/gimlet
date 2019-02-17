package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

///////////////////////////////////////////////////////////////////////////////////
// gmbh Project config
///////////////////////////////////////////////////////////////////////////////////

// SystemConfig - config for gmbh
type SystemConfig struct {
	Core     *SystemCore     `toml:"core"`
	Procm    *SystemProcm    `toml:"procm"`
	Services *SystemServices `toml:"services"`
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

// SystemServices stores information about connecting services
type SystemServices struct {
	ServicesDirectory string   `toml:"services_directory"`
	Services          []string `toml:"services"`
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

// ParseSystemServices returns only the services settings
func ParseSystemServices(configFile string) (*SystemServices, error) {
	system, err := ParseSystemConfig(configFile)
	if err != nil {
		return nil, err
	}
	return system.Services, nil
}

// setDefaults fills in the blanks with default settings on the important config values
func setDefaults(c *SystemConfig) {
	if c.Core.Mode == "" {
		c.Core.Mode = DefaultSystemCore.Mode
	}
	if c.Core.Address == "" {
		c.Core.Address = DefaultSystemCore.Address
	}
	if c.Core.BinPath == "" {
		c.Core.BinPath = DefaultSystemCore.BinPath
	}
	if c.Procm.Address == "" {
		c.Procm.Address = DefaultSystemProcm.Address
	}
	if c.Procm.BinPath == "" {
		c.Procm.BinPath = DefaultSystemProcm.BinPath
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
