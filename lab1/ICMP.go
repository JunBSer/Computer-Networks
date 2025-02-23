package main

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
)

func sendICMPRequest(dstIP, ifaceAddr string) error {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO"),
		},
	}

	data, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	conn, err := icmp.ListenPacket("ip4:icmp", ifaceAddr)
	if err != nil {
		return fmt.Errorf("ошибка создания сокета: %v (запустите от администратора)", err)
	}
	defer conn.Close()

	dst, err := net.ResolveIPAddr("ip4", dstIP)
	if err != nil {
		return err
	}

	if _, err := conn.WriteTo(data, dst); err != nil {
		return err
	}

	return nil
}
