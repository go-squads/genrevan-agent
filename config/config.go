package config

import (
	"fmt"
	"github.com/go-squads/genrevan-agent/util"
	"github.com/spf13/viper"
	"os"
)

var basepath = util.GetRootFolderPath()

func SetupConfig() error {
	viper.SetConfigFile(basepath + "config/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
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
