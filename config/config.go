package config

import (
	"errors"
	"fmt"
	"github.com/go-squads/genrevan-agent/util"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Conf struct {
	SchedulerIp   string `yaml:"SCHEDULER_IP"`
	SchedulerPort string `yaml:"SCHEDULER_PORT"`
	LxdId         string `yaml:"LXD_ID"`
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

func PersistLXDId(id string) {
	f, err := os.OpenFile(basepath+"config/config.yaml", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	if _, err = f.WriteString("LXD_ID: " + id + "\n"); err != nil {
		fmt.Println(err)
	}
}
