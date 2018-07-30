package collector

import (
	"bytes"
	"fmt"
	"github.com/go-squads/genrevan-agent/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"net/http"
	"net/url"
	"strconv"
)

func getCPULoad() string {
	cpuLoad, err := cpu.Percent(0, false)
	if err != nil {
		fmt.Println(err)
	}

	return strconv.FormatFloat(cpuLoad[0], 'f', 3, 64)
}

func getMemoryLoad() string {
	memoryLoad, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println(err)
	}

	return fmt.Sprint(memoryLoad.Used / (1024 * 1024))
}

func SendCurrentLoad() {
	data := url.Values{}
	data.Add("cpu", getCPULoad())
	data.Add("memory", getMemoryLoad())

	client := &http.Client{}
	body := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest(http.MethodPut, "http://"+config.Conf.SchedulerIp+":"+config.Conf.SchedulerPort+"/metric/"+config.Conf.LxdId, body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	respond, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(respond)
}
