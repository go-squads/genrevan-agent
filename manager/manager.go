package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-squads/genrevan-agent/config"
	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
var c *config.Conf

func CheckLXCsState(conf *config.Conf) {
	c = conf
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
			createNewLXC(lxcs[i])
		} else if lxcs[i].Status == "deleted" {
			deleteLXC(lxcs[i])
		} else if lxcs[i].Status == "stopped" {
			updateLXCState(lxcs[i], "stop")
		}
	}
}

func connectToLXD() {
	Lxd, _ = lxd.ConnectLXDUnix("", nil)
}

func createNewLXC(l Lxc) {
	req := api.ContainersPost{
		Name: l.Name,
		Source: api.ContainerSource{
			Type:     "image",
			Protocol: "simplestreams",
			Server:   "https://cloud-images.ubuntu.com/daily",
			Alias:    l.Image,
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

	updateLXCState(l, "start")
}

func updateLXCState(l Lxc, state string) {
	reqState := api.ContainerStatePut{
		Action:  state,
		Timeout: -1,
	}

	op, err := Lxd.UpdateContainerState(l.Name, reqState, "")
	if err != nil {
		fmt.Println(err)
	}

	err = op.Wait()
	if err != nil {
		fmt.Println(err)
	}

	l.Status = "running"

	updateStateToServer(l)
}

func updateStateToServer(l Lxc) {
	form := url.Values{}
	form.Add("state", l.Status)
	body := bytes.NewBufferString(form.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", "http://"+c.SchedulerIp+":"+c.SchedulerPort+"/lxc/"+strconv.Itoa(l.Id)+"/state", body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
}

func deleteLXC(l Lxc) {
	updateLXCState(l, "stop")

	op, err := Lxd.DeleteContainer(l.Name)
	if err != nil {
		fmt.Println(err)
	}

	err = op.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
