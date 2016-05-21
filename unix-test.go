package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	data, err := bufio.NewReader(conn).ReadString('\n')

	defer conn.Close()

	if err != nil {
		log.Print("get client data error: ", err.Error())
	} else {
		fmt.Printf("%#v\n", data)
	}
}

func main() {
	file := "/tmp/socket"

	if _, err := os.Stat(file); err == nil {
		err := os.Remove(file)
		if err != nil {
			panic(err)
		}
	}

	ln, err := net.Listen("unix", file)
	defer os.Remove(file)

	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("get client connection error: ", err)
		} else {
			go handleConnection(conn)
		}
	}
}