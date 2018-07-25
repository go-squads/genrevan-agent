package main

import(
  "github.com/jasonlvhit/gocron"
  "github.com/go-squads/genrevan-agent/collector"
  "fmt"
  "net"
  "net/url"
  "net/http"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "strconv"
  "github.com/go-squads/genrevan-agent/config"
  "os"
)

var Conf *config.Conf

func main() {
  Conf, _ = config.GetConfig()

  if os.Getenv("LXC_ID") == "" {
    register()
  }

  gocron.Every(2).Seconds().Do(collector.SendCurrentLoad, Conf)
  <- gocron.Start()
}

func register() {
  form := url.Values{}
  form.Add("ip", getOutboundIP())
  body := bytes.NewBufferString(form.Encode())
  response, err := http.Post("http://"+Conf.SchedulerIp+":"+Conf.SchedulerPort+"/lxd/register", "application/x-www-form-urlencoded", body)

  if err != nil {
    fmt.Errorf("%v", err)
  }

  responseBody, err := ioutil.ReadAll(response.Body)
  if err != nil {
     fmt.Println(err)
  }

  result := make(map[string]int)

  json.Unmarshal(responseBody, &result)

  os.Setenv("LXD_ID", strconv.Itoa(result["id"]))
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

