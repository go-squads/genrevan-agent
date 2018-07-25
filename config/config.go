package config

import (
	"errors"
	"io/ioutil"

	"github.com/go-squads/genrevan-agent/util"
	yaml "gopkg.in/yaml.v2"
)

type Conf struct {
	SchedulerIp string `yaml:"SCHEDULER_IP"`
	SchedulerPort string `yaml:"SCHEDULER_PORT"`
}

var basepath = util.GetRootFolderPath()

func GetConfig() (*Conf, error) {
	yamlFile, err := ioutil.ReadFile(basepath + "config/config.yaml")
	if err != nil {
		return nil, errors.New("File not found")
	}

	var c *Conf

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
