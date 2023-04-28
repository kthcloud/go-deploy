package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type GpuConfiguration struct {
	PrivilegedCards []string `yaml:"privilegedCards"`
	IgnoredHosts    []string `yaml:"ignoredCards"`
}

var GPU GpuConfiguration

func (g *GpuConfiguration) IsPrivileged(card string) bool {
	for _, c := range g.PrivilegedCards {
		if c == card {
			return true
		}
	}

	return false
}

func (g *GpuConfiguration) IsIgnored(host string) bool {
	for _, c := range g.IgnoredHosts {
		if c == host {
			return true
		}
	}

	return false
}

func SetupGPU() {
	makeError := func(err error) error {
		return fmt.Errorf("failed to setup environment. details: %s", err)
	}

	filepath := Env.GpuConfFilepath

	log.Println("reading gpu config from", filepath)
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf(makeError(err).Error())
	}

	err = yaml.Unmarshal(yamlFile, &GPU)
	if err != nil {
		log.Fatalf(makeError(err).Error())
	}

	log.Println("gpu config loaded")
}
