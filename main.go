package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-squads/genrevan-agent/collector"
	"github.com/go-squads/genrevan-agent/config"
	"github.com/go-squads/genrevan-agent/manager"
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

var Conf *config.Conf

func main() {
	Conf, _ = config.GetConfig()

	if Conf.LxdId == "" {
		register()
		Conf, _ = config.GetConfig()
	}

	managerCJ := gocron.NewScheduler()
	managerCJ.Every(2).Seconds().Do(manager.CheckLXCsState, Conf)
	managerCJ.Start()

	collectorCJ := gocron.NewScheduler()
	collectorCJ.Every(5).Seconds().Do(collector.SendCurrentLoad, Conf)
	<-collectorCJ.Start()
}

func register() {
	form := url.Values{}
	form.Add("ip", getOutboundIP())
	body := bytes.NewBufferString(form.Encode())
	response, err := http.Post("http://"+Conf.SchedulerIp+":"+Conf.SchedulerPort+"/lxd/register", "application/x-www-form-urlencoded", body)

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

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

