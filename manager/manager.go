package manager

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
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
		log.Printf("%v", err)
		return
	}

	var lxcs = []Lxc{}

	bodyBytes, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(bodyBytes, &lxcs)

	for i := 0; i < len(lxcs); i++ {
		switch lxcs[i].Status {
		case "pending":
			createNewLXC(lxcs[i])
			startLXC(lxcs[i])
		case "deleted":
			if isLXCExists(lxcs[i].Name) {
				deleteLXC(lxcs[i])
			}
		case "stopped":
			if isLXCExists(lxcs[i].Name) {
				if isLXCRunning(lxcs[i].Name) {
					log.Printf("%v", "Stopping "+lxcs[i].Name)
					updateLXCState(lxcs[i], "stop")
				}
			} else {
				createNewLXC(lxcs[i])
			}
		case "started":
			if !isLXCExists(lxcs[i].Name) {
				createNewLXC(lxcs[i])
			}
			startLXC(lxcs[i])
		case "running":
			if !isLXCExists(lxcs[i].Name) {
				createNewLXC(lxcs[i])
			}
			if !isLXCRunning(lxcs[i].Name) {
				startLXC(lxcs[i])
			}
		}
	}
}

func connectToLXD() {
	Lxd, _ = lxd.ConnectLXDUnix("", nil)
}

func isLXCExists(name string) bool {
	c := GetContainer(name)

	return c != nil
}

func isLXCRunning(name string) bool {
	c := GetContainer(name)

	return c.Status == "Running"
}

func GetContainer(name string) *api.Container {
	c, _, err := Lxd.GetContainer(name)
	if err != nil {
		return nil
	}

	return c
}

func startLXC(l Lxc) {
	log.Printf("%v", "Starting "+l.Name)
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
		log.Printf("%v", err)
	}

	if op == nil {
		log.Printf("%v", op)
	}

	log.Printf("%v", "Creating "+l.Name+" Container")
	err = op.Wait()
	if err != nil {
		log.Printf("%v", err)
	}
}

func updateLXCState(l Lxc, state string) {
	reqState := api.ContainerStatePut{
		Action:  state,
		Timeout: -1,
	}

	op, err := Lxd.UpdateContainerState(l.Name, reqState, "")
	if err != nil {
		log.Printf("%v", err)
	}

	err = op.Wait()
	if err != nil {
		log.Printf("%v", err)
	}
}

func updateStateToServer(l Lxc) {
	form := url.Values{}
	form.Add("state", l.Status)
	body := bytes.NewBufferString(form.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", "http://"+viper.GetString("SCHEDULER_IP")+":"+viper.GetString("SCHEDULER_PORT")+"/lxc/"+strconv.Itoa(l.Id)+"/state", body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(req)

	if err != nil {
		log.Printf("%v", err)
	}
}

func deleteLXC(l Lxc) {
	log.Printf("%v", "Deleting "+l.Name)
	updateLXCState(l, "stop")

	op, err := Lxd.DeleteContainer(l.Name)
	if err != nil {
		log.Printf("%v", err)
	}

	err = op.Wait()
	if err != nil {
		log.Printf("%v", err)
	}

	deleteLXCFromServer(l)
}

func deleteLXCFromServer(l Lxc) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://"+viper.GetString("SCHEDULER_IP")+":"+viper.GetString("SCHEDULER_PORT")+"/lxc/"+strconv.Itoa(l.Id), nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(req)

	if err != nil {
		log.Printf("%v", err)
	}
}
