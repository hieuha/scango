package main

import (
	"flag"
	"fmt"
	"gopkg.in/redis.v4"
	"log"
	"net"
	"runtime"
	"time"
)

var (
	configPath     string
	Config         config
	rangeIP        string
	logRedis       bool
	rclient        *redis.Client
	payload_ntp_v2 = []byte{0x17, 0x00, 0x03, 0x2a, 0x00, 0x00, 0x00, 0x00}
)

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func ping(pingChan <-chan string, timeOut int) {
	for ip := range pingChan {
		totalByteReceived := 0
		ntpServer, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:123", ip))
		conn, _ := net.DialUDP("udp", nil, ntpServer)
		conn.Write(payload_ntp_v2)
		buf := make([]byte, 1024)
		for {
			conn.SetReadDeadline(time.Now().Add(time.Duration(timeOut) * time.Second))
			len, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if Config.LogLevel > 1 {
					log.Printf("Error:  %s\n", err)
				}
				break
			}
			totalByteReceived += len
		}
		conn.Close()

		//  NTP Server alive
		if totalByteReceived > 0 {
			if Config.LogLevel > 0 {
				log.Printf("NTP Server:  %s\n", fmt.Sprintf(" %s len %d", ip, totalByteReceived))
			}
			if logRedis {
				time_now := time.Now().UnixNano()
				if rclient != nil {
					rclient.LPush("NTP_SERVER", time_now, ip)
				}
			}
		}
	}
}

func init() {
	flag.StringVar(&configPath, "config", CONFIG_FILE, "location of the config file")
	flag.StringVar(&rangeIP, "range", "", "Range IP for scanning")
	flag.BoolVar(&logRedis, "redis", false, "Enable redis for logging")
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	if err := LoadConfig(configPath); err != nil {
		log.Fatal(err)
	}

	coreLogFile, err := LoggerInit(Config.CoreLog)
	if err != nil {
		log.Fatal(err)
	}
	defer coreLogFile.Close()

	// Init redis server
	if logRedis {
		rclient = redis.NewClient(&redis.Options{
			Addr:     Config.Rediserver,
			Password: Config.Redisauth,
			DB:       Config.Redisdb,
		})
	}

	hosts, _ := Hosts(rangeIP)
	totalHosts := len(hosts)
	if Config.LogLevel > 0 {
		log.Printf("Total host %d\n", totalHosts)
	}

	pingChan := make(chan string, Config.Concurrentmax)

	// Create threats for checking
	for i := 0; i < Config.Concurrentmax; i++ {
		go ping(pingChan, Config.Timeout)
	}

	// Add a host of list to chan for scanning
	for _, ip := range hosts {
		pingChan <- ip
	}
}
