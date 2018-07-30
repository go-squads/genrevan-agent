package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"github.com/spf13/viper"
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

func CheckLXCsState() {
	connectToLXD()

	response, err := http.Get("http://" + viper.GetString("SCHEDULER_IP") + ":" + viper.GetString("SCHEDULER_PORT") + "/lxc/lxd/" + viper.GetString("LXD_ID"))
	if err != nil {
		fmt.Println(err)
		return
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
		} else if lxcs[i].Status == "started" {
			startLXC(lxcs[i])
		} else if lxcs[i].Status == "running" {
			if !isLXCExists(lxcs[i].Name) {
				createNewLXC(lxcs[i])
			}
		}
	}
}

func connectToLXD() {
	Lxd, _ = lxd.ConnectLXDUnix("", nil)
}

func isLXCExists(name string) bool {
	_, _, err := Lxd.GetContainer(name)
	if err != nil {
		return false
	}

	return true
}

func startLXC(l Lxc) {
	updateLXCState(l, "start")

	l.Status = "running"
	updateStateToServer(l)
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

	startLXC(l)
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
}

func updateStateToServer(l Lxc) {
	form := url.Values{}
	form.Add("state", l.Status)
	body := bytes.NewBufferString(form.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", "http://"+ viper.GetString("SCHEDULER_IP") +":"+ viper.GetString("SCHEDULER_PORT") +"/lxc/"+ strconv.Itoa(l.Id)+"/state", body)
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
