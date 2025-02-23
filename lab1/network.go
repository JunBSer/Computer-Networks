package main

import (
	"errors"
	"log"
	"net"
)

func findInterfaces() []net.Interface {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Network interfaces could not be obtained: %v", err)
	}
	return interfaces
}

func getSubNetworkCIDR(iFaces net.Interface) (*net.IPNet, error) {
	addrs, err := iFaces.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
			return ipNet, nil
		}
	}

	return nil, errors.New("No IP address found")
}

func getIpLimits(baseNet *net.IPNet) (lowerL net.IP, upperL net.IP) {
	baseMask := baseNet.Mask
	baseIP := baseNet.IP.To4()

	if baseIP == nil {
		return nil, nil
	}

	lower := make(net.IP, len(baseIP))
	upper := make(net.IP, len(baseIP))

	copy(lower, baseIP)
	copy(upper, baseIP)

	for i := range baseMask {
		lower[i] &= baseMask[i]
		upper[i] |= ^baseMask[i]
	}

	return lower, upper // Возвращаем значения, а не указатели
}

func getIPRange(lIP, rIP *net.IP) []net.IP {
	var ips []net.IP

	for ip := *lIP; !ip.Equal(*rIP); ip = incrementIp(ip) {
		ips = append(ips, ip)
	}
	ips = append(ips, *rIP)
	return ips
}

func incrementIp(ip net.IP) net.IP {
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)

	for i := len(newIP) - 1; i >= 0; i-- {
		newIP[i]++
		if newIP[i] != 0 {
			break
		}
	}
	return newIP
}
