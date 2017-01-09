# scango
```
GOARCH=amd64 GOOS=linux go build
```
```
Usage of ./scango:
  -config string
        location of the config file (default "./config.toml")
  -range string
        Range IP for scanning
  -redis
        Enable redis for logging
```
```
redis-cli -a deptraivair --csv lrange NTP_SERVER 0 -1 |grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}' |sort|uniq > ip.txt
```
