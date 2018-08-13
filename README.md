# GENREVAN AGENT
Genrevan Agent is an open source project to consume Genrevan Scheduler services. The main functions of this applicatiion are:
- Send load metric to Genrevan Scheduler
- Check for new LXC to be allocate

## Dependencies
- [shirou/gopsutil](https://github.com/shirou/gopsutil)
- [robfig/cron](https://github.com/robfig/cron)
- netfilter-persistent

## Build
```go install genrevan-agent```

## Run
```genrevan-agent```

## Test
```go test -v```

