package static

import (
	"errors"
	"io/ioutil"

	"github.com/gmbh-micro/defaults"
	yaml "gopkg.in/yaml.v2"
)

// Static is the parsable config for individual services
type Static struct {
	Name         string   `yaml:"name"`
	Aliases      []string `yaml:"aliases"`
	Language     string   `yaml:"language"`
	MakefilePath string   `yaml:"path_to_makefile"`
	BinPath      string   `yaml:"path_to_bin"`
	SrcPath      string   `yaml:"path_to_src"`
	IsClient     bool     `yaml:"is_client"`
	IsServer     bool     `yaml:"is_server"`
}

// ParseData attempts to parse a yaml file at path and return the UserConfigStruct.
// If not all settings have been defined in user path, the defaults will be used.
func ParseData(path string) (*Static, error) {
	config := Static{IsClient: defaults.IS_CLIENT, IsServer: defaults.IS_SERVER}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("could not open yaml file: " + err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, errors.New("could not parse yaml file: " + err.Error())
	}

	setDefaultStatic(&config)

	return &config, nil
}

func setDefaultStatic(c *Static) {
	if c.Name == "" {
		c.Name = defaults.SERVICE_NAME
	}
}

// DataIsValid returns true if the Static object has the required information
func DataIsValid(c *Static) bool {
	if c.Language == "" {
		// notify.StdMsgDebug("invalid lang")
		return false
	}
	if c.BinPath == "" && c.SrcPath == "" {
		// notify.StdMsgDebug("invalid path")
		return false
	}
	if c.Name == defaults.SERVICE_NAME {
		// notify.StdMsgDebug("invalid service name")
		return false
	}
	return true
}
