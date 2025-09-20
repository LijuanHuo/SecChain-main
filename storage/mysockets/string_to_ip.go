package mysockets

import (
	"log"
	"net"
	"strconv"
	"strings"
)

func String_to_IP(s string) net.TCPAddr {
	ss := strings.Split(s, ":")
	num, err := strconv.Atoi(ss[1])
	if err != nil {
		log.Fatal(err)
	}

	ipAddr := net.TCPAddr{
		IP:   net.ParseIP(ss[0]),
		Port: num,
	}
	return ipAddr
}
