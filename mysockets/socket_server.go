package mysockets

import (
	"fmt"
	"log"
	"net"
)

const TCP_BUFF_SIZE = 1024

func Server(serverIP string) {
	listen, err := net.Listen("tcp", serverIP)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("con=%v , ip=%v\n", conn, conn.RemoteAddr().String())

		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	fmt.Printf("waitting for %s\n", c.RemoteAddr().String())
	buf := make([]byte, TCP_BUFF_SIZE)
	n, err := c.Read(buf)
	if err != nil {
		log.Fatal()
	}

	fmt.Print(string(buf[:n]))

	str := "hello QwQ"

	len, err := c.Write([]byte(str + "\n"))

	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(len) + "\n")

}
