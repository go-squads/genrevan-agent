package config

import (
	"errors"
	"io/ioutil"

	"github.com/go-squads/genrevan-agent/util"
	yaml "gopkg.in/yaml.v2"
)

type Conf struct {
	SchedulerIP string `yaml:"SCHEDULER_IP"`
	SchedulerPort string `yaml:"SCHEDULER_PORT"`
	LxdId	int	`yaml:"LXD_ID"`
}

var basepath = util.GetRootFolderPath()

func (c *Conf) GetConfig() (*Conf, error) {
	yamlFile, err := ioutil.ReadFile(basepath + "config/config.yaml")
	if err != nil {
		return nil, errors.New("File not found")
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
