package main

import (
	"fmt"
	"io"
	"net"
)

// this is a demo 1to1 server, it will accept one conn and handle it,
// if you want to handle multi conn, you can use a loop to accept conn,
// and use a map to store conn, key is conn id, value is conn.
func main() {

	startListener()

}

func startListener() {
	var err error
	var listener net.Listener
	listener, err = net.Listen("tcp", "127.0.0.1:8080")
	if err == nil {
		fmt.Println("listen success")
	}
	defer listener.Close()

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err == nil {
			fmt.Printf("accept conn success, conn: %v", conn)
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("read data success, data: %s", string(buf[:n]))
			} else {
				fmt.Printf("read data failed, err: %v", err)
			}
			break
		}

		data := buf[:n]
		fmt.Printf("read data success, data: %s", string(data))

		response := processRequest(data)

		_, err = conn.Write([]byte(response))
		if err == nil {
			fmt.Printf("write data success, data: %s", string(response))
		}
	}

}

func processRequest(data []byte) []byte {
	var data2 []byte = []byte("learn1!")
	return data2
}
