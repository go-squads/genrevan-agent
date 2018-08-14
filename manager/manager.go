package manager

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-squads/genrevan-agent/iptables"
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
	HostPort  int    `json:"host_port"`
	ContainerPort int `json:"container_port"`
}

var Lxd lxd.ContainerServer

func CheckLXCsStateFromServer() {
	connectToLXD()

	response, err := http.Get("http://" + viper.GetString("SCHEDULER_IP") + ":" + viper.GetString("SCHEDULER_PORT") + "/lxc/lxd/" + viper.GetString("LXD_ID"))
	if err != nil {
		log.Printf("%v", err)
		return
	}

	var lxcs = []Lxc{}

	bodyBytes, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(bodyBytes, &lxcs)

	done := make(chan bool)

	for i := 0; i < len(lxcs); i++ {
		go checkLXCState(lxcs[i], done)
	}

	for i := 0; i < len(lxcs); i++ {
		<-done
	}
}

func checkLXCState(lxc Lxc, done chan bool) {
	switch lxc.Status {
	case "pending":
		createNewLXC(lxc)
		startLXC(lxc)
	case "deleted":
		if isLXCExists(lxc.Name) {
			deleteLXC(lxc)
		}
	case "stopped":
		if isLXCExists(lxc.Name) {
			if isLXCRunning(lxc.Name) {
				log.Printf("%v", "Stopping "+lxc.Name)
				updateLXCState(lxc, "stop")
			}
		} else {
			createNewLXC(lxc)
		}
	case "started":
		if !isLXCExists(lxc.Name) {
			createNewLXC(lxc)
		}
		startLXC(lxc)
	case "running":
		if !isLXCExists(lxc.Name) {
			createNewLXC(lxc)
		}
		if !isLXCRunning(lxc.Name) {
			startLXC(lxc)
		}
	}
	done <- true
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

func getContainerAddress(name string) string {
	state, _, err := Lxd.GetContainerState(name)
	if err != nil {
		log.Printf("%v", err)
	}

	addresses := state.Network["eth0"].Addresses
	for _, address := range addresses {
		if address.Family == "inet" {
			return address.Address
		}
	}

	return ""
}

func registerContainerAddress(l Lxc) {
	var address string
	for {
		address = getContainerAddress(l.Name)
		if len(address) > 0 {
			break
		}
	}

	l.IpAddress = address

	rule := iptables.Rule{
		SourceIP:        viper.GetString("LXD_IP"),
		SourcePort:      strconv.Itoa(l.HostPort),
		DestinationIP:   l.IpAddress,
		DestinationPort: strconv.Itoa(l.ContainerPort),
	}

	updateLXCIPToServer(l)

	err := iptables.Insert(rule)
	if err != nil {
		log.Println(err)
	}

	err = iptables.Save()
	if err != nil {
		log.Println(err)
	}
}

func updateLXCIPToServer(l Lxc) {
	form := url.Values{}
	form.Add("ip", l.IpAddress)
	body := bytes.NewBufferString(form.Encode())

	httpClient := &http.Client{}
	url := "http://"+viper.GetString("SCHEDULER_IP")+":"+viper.GetString("SCHEDULER_PORT")+"/lxc/"+strconv.Itoa(l.Id)+"/ip"
	req, err := http.NewRequest(http.MethodPatch, url, body)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		log.Println("%v", err)
	}

	httpClient.Do(req)

	if err != nil {
		log.Println("%v", err)
	}
}

func startLXC(l Lxc) {
	log.Printf("%v", "Starting "+l.Name)
	updateLXCState(l, "start")

	if isLXCRunning(l.Name) {
		l.Status = "running"
		updateStateToServer(l)
		go registerContainerAddress(l)
	}
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
