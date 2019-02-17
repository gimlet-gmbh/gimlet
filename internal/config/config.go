package config

import (
	"errors"
	"time"

	"github.com/BurntSushi/toml"
)

///////////////////////////////////////////////////////////////////////////////////
// gmbh Project config
///////////////////////////////////////////////////////////////////////////////////
type Core struct {
	Settings *Settings    `toml:"core"`
	Procm    *Procm       `toml:"procm"`
	Service  *CoreService `toml:"services"`
}

type Settings struct {
	Mode      string   `toml:"mode"`
	Address   string   `toml:"address"`
	KeepAlive duration `toml:"keep_alive"`
	Verbose   bool     `toml:"verbose"`
	Daemon    bool     `toml:"daemon"`
}

type Procm struct {
	Address   string   `toml:"address"`
	KeepAlive duration `toml:"keep_alive"`
	Verbose   bool     `toml:"verbose"`
}

type CoreService struct {
	ServicesDirectory string   `toml:"services_directory"`
	Services          []string `toml:"services"`
}

func ParseCore(path string) (*Core, error) {
	var gmbh Core
	if _, err := toml.DecodeFile(path, &gmbh); err != nil {
		return nil, err
	}
	return &gmbh, nil
}

func ParseProcm(path string) (*Procm, error) {
	g, e := ParseCore(path)
	if e != nil {
		return nil, e
	}
	return g.Procm, nil
}

func ParseSettings(path string) (*Settings, error) {
	g, e := ParseCore(path)
	if e != nil {
		return nil, e
	}
	return g.Settings, nil
}

func ParseCoreService(path string) (*CoreService, error) {
	g, e := ParseCore(path)
	if e != nil {
		return nil, e
	}
	return g.Service, nil
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

///////////////////////////////////////////////////////////////////////////////////
// gmbh Service config
///////////////////////////////////////////////////////////////////////////////////

type Service struct {
	Mode       string  `toml:"mode"`
	StaticData *Static `toml:"connecting"`
}

type Static struct {
	Name     string   `toml:"name"`
	Aliases  []string `toml:"aliases"`
	Args     []string `toml:"args"`
	Env      []string `toml:"env"`
	Language string   `toml:"language"`
	BinPath  string   `toml:"bin_path"`
	SrcPath  string   `toml:"src_path"`
}

func ParseService(path string) (*Service, error) {
	var service Service
	if _, err := toml.DecodeFile(path, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

func ParseStaticService(path string) (*Static, error) {
	g, e := ParseService(path)
	if e != nil {
		return nil, e
	}
	return g.StaticData, nil
}

func (s *Static) Validate() error {
	if s.Name == "" {
		return errors.New("empty name")
	}
	oneOrOther := false
	if s.BinPath != "" {
		oneOrOther = true
	}
	if s.Language != "" && s.SrcPath != "" {
		oneOrOther = true
	}
	if !oneOrOther {
		return errors.New("unclear ")
	}
	return nil
}
