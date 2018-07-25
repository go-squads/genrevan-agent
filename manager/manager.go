package manager

import (
	"encoding/json"
	"fmt"
	"github.com/go-squads/genrevan-agent/config"
	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"io/ioutil"
	"net/http"
	"os"
)

type Lxc struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	IpAddress string `json:"ip_address"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	LxdId     int    `json:"lxd_id"`
}

var Lxd lxd.ContainerServer

func CheckLXCsState(conf *config.Conf) {
	connectToLXD()

	response, err := http.Get("http://" + conf.SchedulerIp + ":" + conf.SchedulerPort + "/lxc/lxd/" + os.Getenv("LXD_ID"))
	if err != nil {
		fmt.Println(err)
	}

	var lxcs = []Lxc{}

	bodyBytes, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(bodyBytes, &lxcs)

	for i := 0; i < len(lxcs); i++ {
		if lxcs[i].Status == "pending" {
			createNewLXC(lxcs[i].Name, lxcs[i].Image)
		}
	}
}

func connectToLXD() {
	Lxd, _ = lxd.ConnectLXDUnix("", nil)
}

func createNewLXC(name, image string) {
	req := api.ContainersPost{
		Name: name,
		Source: api.ContainerSource{
			Type:     "image",
			Protocol: "simplestreams",
			Server:   "https://cloud-images.ubuntu.com/daily",
			Alias:    image,
		},
	}

	op, err := Lxd.CreateContainer(req)
	if err != nil {
		fmt.Println(err)
	}

	if op == nil {
		fmt.Println(op)
	}

	err = op.Wait()
	if err != nil {
		fmt.Println(err)
	}

	updateLXCState(name, "start")
}

func updateLXCState(name, state string) {
	reqState := api.ContainerStatePut{
		Action:  state,
		Timeout: -1,
	}

	op, err := Lxd.UpdateContainerState(name, reqState, "")
	if err != nil {
		fmt.Println(err)
	}

	err = op.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
