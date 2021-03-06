package main

import (
	"net"
	"strconv"
	"time"
	"log"
	"sync"
	"sync/atomic"
)

var SOCKET string = "/tmp/log-agent.socket"

func main() {
	var totalNum int32 = 0
	errorNum := 0
	lenght := 64

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
			total := 20000000
			json := "{\"test\":" + strconv.FormatInt(time.Now().Unix(), 10) + ",\"hello world\":{\"time\":11133322}}\n"
			//json := strconv.FormatInt(time.Now().Unix() / 10, 10) + "\n"
			for i := 1; i <= total; i ++ {
				conn.Write([]byte(json))
				atomic.AddInt32(&totalNum, 1)
			}
		}()
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	log.Println("total: ", totalNum)
	log.Println("error: ", errorNum)
	log.Printf("run time: %.2f ms", float64(time.Now().UnixNano() - startime) / 1000 / 1000)
}
