package main
import (
    "fmt"
    "net"
    "time"
)

var (
    PAYLOAD_NTP_V2 = []byte{0x17, 0x00, 0x03, 0x2a, 0x00, 0x00, 0x00, 0x00}
    TIME_OUT time.Duration = 5
    CONCURRENT_MAX = 200
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

func ping(pingChan <-chan string, timeOut time.Duration) {
    for ip := range pingChan {
        totalByteReceived := 0
        ntpServer, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:123", ip))
        conn, _ := net.DialUDP("udp", nil, ntpServer)
        conn.Write(PAYLOAD_NTP_V2)
        buf := make([]byte, 1024)
        for {
            conn.SetReadDeadline(time.Now().Add(timeOut * time.Second))
            len, _, err := conn.ReadFromUDP(buf)
            if err != nil {
                // fmt.Println("Error: ",err)
                break
            }
            totalByteReceived += len
        }
        conn.Close()

       //  NTP Server alive
       if totalByteReceived > 0 {
            fmt.Println(fmt.Sprintf("recv UDP from %s len %d", ip, totalByteReceived))
        }
    }
}

func main() {
    hosts, _ := Hosts("1.52.0.0/14")
    totalHosts := len(hosts)
    fmt.Println("Total IP", totalHosts)
    pingChan := make(chan string, CONCURRENT_MAX)

    // Create threats for checking
    for i := 0; i < CONCURRENT_MAX; i++ {
        go ping(pingChan, TIME_OUT)
    }

    // Add a host of list to chan for scanning
    for _, ip := range hosts {
        pingChan <- ip
    }
}