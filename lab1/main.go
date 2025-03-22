package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const maxWorkers = 200

func findOpenedPorts(ipAddr string) []string {
	var openedPorts []string
	var mutex sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers)

	for port := 1; port <= 2000; port++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			target := fmt.Sprintf("%s:%d", ipAddr, p)
			conn, err := net.DialTimeout("tcp", target, 1*time.Second)
			if err != nil {
				return
			}
			conn.Close()

			mutex.Lock()
			openedPorts = append(openedPorts, strconv.Itoa(p))
			mutex.Unlock()

			time.Sleep(10 * time.Millisecond)
		}(port)
	}

	wg.Wait()
	return openedPorts
}

func main() {
	iFaces := findInterfaces()

	mainIFace, err := ChooseInterface(iFaces)
	if err != nil {
		panic(err)
	}

	iFaceNet, err := getSubNetworkCIDR(mainIFace)

	lIP, rIP := getIpLimits(iFaceNet)

	ipAddrs := getIPRange(&lIP, &rIP)

	wg := sync.WaitGroup{}
	sem := make(chan struct{}, maxWorkers)

	for i := 1; i < len(ipAddrs); i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			sendICMPRequest(ipAddrs[i].String(), iFaceNet.IP.To4().String())
			time.Sleep(time.Millisecond * 500)
		}()
	}
	wg.Wait()

	time.Sleep(5 * time.Second)

	parsedArp, err := ParseARPTable(iFaceNet.IP.To4().String())
	if err != nil {
		panic(err)
	}

	for key, val := range parsedArp {
		fmt.Printf("IP: %s   Mac: %s\n", key, val)
		fmt.Println(findOpenedPorts(key))
	}

}
