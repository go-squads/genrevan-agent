package collector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/viper"
)

func getCPULoad() string {
	cpuLoad, err := cpu.Percent(0, false)
	if err != nil {
		log.Println(err)
	}

	return strconv.FormatFloat(cpuLoad[0], 'f', 3, 64)
}

func getMemoryLoad() string {
	memoryLoad, err := mem.VirtualMemory()
	if err != nil {
		log.Println(err)
	}

	return fmt.Sprint(memoryLoad.Used / (1024 * 1024))
}

func SendCurrentLoad() {
	data := url.Values{}
	data.Add("cpu", getCPULoad())
	data.Add("memory", getMemoryLoad())

	client := &http.Client{}
	body := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest(http.MethodPut, "http://"+viper.GetString("SCHEDULER_IP")+":"+viper.GetString("SCHEDULER_PORT")+"/metric/"+viper.GetString("LXD_ID"), body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)

	if err != nil {
		log.Println(err)
	}

	resBody, _ := ioutil.ReadAll(response.Body)
	log.Println("Send Metrics " + string(resBody))
}
