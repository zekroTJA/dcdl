package config

import (
	"os"

	"github.com/traefik/paerser/env"
	"github.com/traefik/paerser/file"
	"github.com/traefik/paerser/flag"
	"github.com/zekrotja/dcdl/models"
)

const defaultConfigLoc = "./config.yaml"

type Paerser struct {
	cfg        *models.ConfigModel
	args       []string
	configFile string
}

var _ ConfigProvider = (*Paerser)(nil)

func NewPaerser(args []string, configFile string) *Paerser {
	return &Paerser{
		args:       args,
		configFile: configFile,
	}
}

func (p *Paerser) Instance() *models.ConfigModel {
	return p.cfg
}

func (p *Paerser) Load() (err error) {
	cfg := models.DefaultConfig

	cfgFile := defaultConfigLoc
	if p.configFile != "" {
		cfgFile = p.configFile
	}
	if err = file.Decode(cfgFile, &cfg); err != nil && !os.IsNotExist(err) {
		return
	}

	if err = env.Decode(os.Environ(), "DCDL_", &cfg); err != nil {
		return
	}

	args := os.Args[1:]
	if p.args != nil {
		args = p.args
	}
	if err = flag.Decode(args, &cfg); err != nil {
		return
	}

	p.cfg = &cfg

	return
}
