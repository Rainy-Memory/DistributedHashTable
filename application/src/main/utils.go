package main

import (
	"net"
)

func GetLocalAddress() string {
	var localAddress string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localAddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localAddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localAddress
}

func appHash(org string) string {
	return "*(&*(^" + org + "%^%*)(*%)*&"
}
