package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	var conn net.Conn
	var err error
	conn, err = startListener()
	if err == nil {
		handleConnection(conn)
	}
}

func startListener() (net.Conn, error) {
	var err error
	var listener net.Listener
	listener, err = net.Listen("tcp", "127.0.0.1:8080")
	if err == nil {
		fmt.Println("listen success")
	}
	defer listener.Close()

	var conn net.Conn
	conn, err = listener.Accept()
	if err == nil {
		fmt.Printf("accept conn success, conn: %v", conn)
	}
	return conn, nil
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
