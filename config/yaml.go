package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	FileLocation string
}

func (conf *YamlConfig) GetAllFields() (*CasaConfig, error) {
	content, err := ioutil.ReadFile(conf.FileLocation)
	if err != nil {
		return nil, err
	}
	var MiCasaConfig CasaConfig
	err = yaml.Unmarshal(content, &MiCasaConfig)
	if err != nil {
		return nil, err
	}
	return &MiCasaConfig, nil
}
