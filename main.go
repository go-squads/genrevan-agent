package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-squads/genrevan-agent/collector"
	"github.com/go-squads/genrevan-agent/config"
	"github.com/go-squads/genrevan-agent/manager"
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"net/http"
	"strconv"
	"github.com/spf13/viper"
)

func main() {
	setupConfiguration()

	if viper.Get("LXD_ID") == "" {
		register()
		setupConfiguration()
	}

	managerCJ := gocron.NewScheduler()
	managerCJ.Every(2).Seconds().Do(manager.CheckLXCsState)
	managerCJ.Start()

	collectorCJ := gocron.NewScheduler()
	collectorCJ.Every(5).Seconds().Do(collector.SendCurrentLoad)
	<-collectorCJ.Start()
}

func setupConfiguration() {
	err := config.SetupConfig()
	if err != nil {
		fmt.Println(err)
	}
}

func register() {
	response, err := http.Get("http://"+ viper.GetString("SCHEDULER_IP")  +":"+ viper.GetString("SCHEDULER_PORT") +"/lxd/register")

	if err != nil {
		fmt.Println(err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	result := make(map[string]int)

	json.Unmarshal(responseBody, &result)

	config.PersistLXDId(strconv.Itoa(result["id"]))
}

