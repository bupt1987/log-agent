package main

import (
	"net"
	"strconv"
	"time"
	"log"
	"sync"
)

var SOCKET string = "/tmp/go-unix.socket"

func main() {
	errorNum := 0
	lenght := 1000

	startime := time.Now().Unix()

	var wg sync.WaitGroup

	for j := 1; j <= lenght; j ++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var conn net.Conn
			var err error

			conn, err = net.Dial("unix", SOCKET)

			if err != nil {
				errorNum++
				log.Println(err.Error())
				return
			}

			defer conn.Close()
			total := 100
			for i := 1; i <= total; i ++ {
				conn.Write([]byte(strconv.Itoa(i) + "\n"))
			}
			sleep := time.Duration(1000 / total)
			time.Sleep(time.Millisecond * sleep)
		}()
	}

	wg.Wait()

	log.Println("error: ", errorNum)
	log.Println("run time: ", time.Now().Unix() - startime, "s")
}
