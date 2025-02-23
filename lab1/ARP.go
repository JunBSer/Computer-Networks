package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

var macPattern = regexp.MustCompile(`^([0-9A-Fa-f]{2}-){5}[0-9A-Fa-f]{2}$`)

func ParseARPTable(interfaceIP string) (map[string]string, error) {
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute arp -a: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	arpMap := make(map[string]string)
	foundInterface := false
	parsing := false
	skipHeader := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if !foundInterface {
			if strings.Contains(line, interfaceIP) {
				foundInterface = true
				parsing = true
				skipHeader = true
			}
			continue
		}

		if parsing {
			if skipHeader {
				skipHeader = false
				continue
			}

			if line == "" {
				parsing = false
				break
			}

			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip, mac := parts[0], parts[1]

				if isValidIP(ip) && isValidMAC(mac) {
					arpMap[ip] = mac
				}
			}
		}
	}

	if !foundInterface {
		return nil, fmt.Errorf("interface with IP %s not found", interfaceIP)
	}

	return arpMap, nil
}

func isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}

func isValidMAC(mac string) bool {
	return macPattern.MatchString(mac)
}
