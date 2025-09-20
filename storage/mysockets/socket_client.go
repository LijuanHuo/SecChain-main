package mysockets

import (
	"fmt"
	"log"
	"net"
)

func StrClient(ip0 string, ip1 string, message string) string {
	var (
		conn net.Conn
		err  error
	)

	remoteAddr := String_to_IP(ip1)
	if ip0 != "" {
		localAddr := String_to_IP(ip0)
		conn, err = net.DialTCP("tcp", &localAddr, &remoteAddr)
	} else {
		conn, err = net.DialTCP("tcp", nil, &remoteAddr)
	}
	if err != nil {
		log.Fatal(err)
	}

	conent, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("send complete: %d bytes\n", conent)

	buf := make([]byte, TCP_BUFF_SIZE)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	var sstr string = string(buf[:n])
	fmt.Print("recive: " + sstr + "\n")

	conn.Close()

	return sstr
}
