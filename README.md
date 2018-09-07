# GENREVAN AGENT
Genrevan Agent is an open source project to consume Genrevan Scheduler services. The main functions of this application are:
- Send load metric to Genrevan Scheduler
- Check for new LXC to be allocate

## Development Environment
- Golang v 1.10.3

## Dependencies
- [shirou/gopsutil](https://github.com/shirou/gopsutil)
- [jasonlvhit/gocron](github.com/jasonlvhit/gocron)
- [netfilter-persistent](http://manpages.ubuntu.com/manpages/xenial/man8/netfilter-persistent.8.html)
- [spf13/viper](https://github.com/spf13/viper)
- [lxc/lxd](https://github.com/lxc/lxd)

## Configuration
Copy config.example.yaml as config.yaml
- Scheduler_IP and Scheduler_Port is for Genrevan Scheduler Address
- LXD_IP is IP Address where Genrevan Agent Located

## Build
```go install genrevan-agent```

## Run
```genrevan-agent```

## Test
```go test -v```

## Additional Notes
If you installing LXD from snap (using LXD version 3 and newer), you should add LXD_SOCKET to enviroment variable
```LXD_SOCKET=/var/snap/lxd/common/lxd/unix.socket```
