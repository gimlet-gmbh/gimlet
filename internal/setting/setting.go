package setting

import (
	"errors"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/gmbh-micro/defaults"
)

// UserConfig represents the parsable config settings
type UserConfig struct {
	Name              string   `yaml:"project_name"`
	Verbose           bool     `yaml:"verbose"`
	Daemon            bool     `yaml:"daemon"`
	DefaultHost       string   `yaml:"default_host"`
	DefaultPort       string   `yaml:"default_port"`
	ControlHost       string   `yaml:"control_host"`
	ControlPort       string   `yaml:"control_port"`
	ServicesDirectory string   `yaml:"services_directory"`
	ServicesToAttach  []string `yaml:"services_to_attach"`
	ServicesDetached  []string `yaml:"services_detached"`
}

// ParseUserConfig attempts to parse a yaml file at path and return the UserConfigStruct.
// If not all settings have been defined in user path, the defaults will be used.
func ParseUserConfig(path string) (*UserConfig, error) {
	config := UserConfig{Verbose: defaults.VERBOSE, Daemon: defaults.DAEMON}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("could not open yaml file: " + err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, errors.New("could not parse yaml file: " + err.Error())
	}

	setDefaultUserConfig(&config)

	return &config, nil
}

func setDefaultUserConfig(c *UserConfig) {
	if c.Name == "" {
		c.Name = defaults.PROJECT_NAME
	}

	if c.DefaultHost == "" {
		c.DefaultHost = defaults.DEFAULT_HOST
	}

	if c.DefaultPort == "" {
		c.DefaultPort = defaults.DEFAULT_PORT
	}

	if c.ControlHost == "" {
		c.ControlHost = defaults.CONTROL_HOST
	}

	if c.ControlPort == "" {
		c.ControlPort = defaults.CONTROL_PORT
	}
}
