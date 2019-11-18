package config

import "github.com/oskoss/mi-casa/thermostat"

type Configurator interface {
	GetAllFields() (config CasaConfig, err error)
}

type CasaConfig struct {
	Name             string                        `yaml:"name"`
	DysonHotCoolLink []thermostat.DysonHotCoolLink `yaml:"dysonHotCoolLinkDevices"`
}
