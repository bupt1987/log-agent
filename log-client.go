package main

import (
	"net"
	"strconv"
	"time"
	"log"
	"sync"
	"sync/atomic"
)

var SOCKET string = "/tmp/go-unix.socket"

func main() {
	var totalNum int32 = 0
	errorNum := 0
	lenght := 256

	startime := time.Now().UnixNano()

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
			total := 100000
			for i := 1; i <= total; i ++ {
				conn.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10) + "\n"))
				atomic.AddInt32(&totalNum, 1)
			}
		}()
	}
	wg.Wait()
	log.Println("total: ", totalNum)
	log.Println("error: ", errorNum)
	log.Printf("run time: %.2f ms", float64(time.Now().UnixNano() - startime) / 1000 / 1000)
}
