package main

import (
	"net"
	"fmt"
	"strconv"
	"time"
	"log"
)

var SOCKET string = "/tmp/go-unix.socket"

func main() {
	errorNum := 0
	count := 0
	lenght := 100
	golist := make(chan int, lenght)
	exit := make(chan int, 1)

	for j := 1; j <= lenght; j ++ {
		golist <- j
	}
	for {
		select {
		case index := <-golist:
			go func() {
				defer func() {
					exit <- 1
				}()
				conn, err := net.DialTimeout("unix", SOCKET, time.Second * 3)
				if err != nil {
					errorNum++
					fmt.Println(err.Error())
					return
				}
				defer conn.Close()

				for i := 1; i <= 10000; i ++ {
					fmt.Fprint(conn, strconv.Itoa(i) + "\n")
				}

				fmt.Println(index)
			}()
		case <-exit:
			count++
			if count >= lenght {
				log.Print("error: ", errorNum)
				return
			}
		}
	}
}
