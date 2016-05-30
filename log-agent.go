package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"log"
	"io"
	"./safe"
	"runtime"
	"bytes"
	"./logger"
	"time"
	"flag"
	"fmt"
	"math/rand"
)

/**
 * for test: -to=http://localhost:8080/upload -type=http
 */

func main() {
	socket := flag.String("socket", "/tmp/log-agent.socket", "Listen socket address")
	sType := flag.String("type", "file", "Process Log type: file or http")
	toAddress := flag.String("to", "/go/logs/", "Log store to address\n\ttype is file, it is local file address\n\ttype is http, it is http address")
	iMaxSize := flag.Int("size", 1024000, "One log max byte")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	iNum := runtime.NumCPU()
	lQueue := make([]*safe.Queue, iNum)
	chConn := make(chan net.Conn, iNum)
	chSig := make(chan os.Signal)
	signal.Notify(chSig, os.Interrupt)
	signal.Notify(chSig, os.Kill)
	signal.Notify(chSig, syscall.SIGTERM)

	//监听
	listen, err := net.Listen("unix", *socket)
	if err != nil {
		panic(err)
	}

	if err := os.Chmod(*socket, 0777); err != nil {
		panic(err)
	}

	for i := 0; i < iNum; i ++ {
		lQueue[i] = safe.NewQueue()
	}

	for n := 0; n < iNum; n ++ {
		go func(n int, lQueue []*safe.Queue) {
			writer := logger.NewLogger(*sType, *toAddress)
			for {
				data := lQueue[n].Pop();
				if data != nil {
					writer.Write(logger.NewPack(n, data.([]byte)))
				} else {
					time.Sleep(time.Second)
				}
			}
		}(n, lQueue)
	}

	go func(listen net.Listener) {
		for {
			conn, err := listen.Accept()
			if err != nil {
				log.Println("connection error:", err)
				continue
			}
			chConn <- conn
		}
	}(listen)

	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("%8s = %v", f.Name, f.Value.String())
	})

	for {
		select {
		case <-chSig:
			if err := os.Remove(*socket); err != nil {
				panic(err)
			}
			return
		case conn := <-chConn:
			go func(iNum int) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				var buffer bytes.Buffer
				for {
					data, err := reader.ReadBytes('\n')
					if len(data) > 0 {
						buffer.Write(data)
					}
					l := buffer.Len()
					if l >= *iMaxSize || (err == io.EOF && l > 0) {
						lQueue[r.Intn(iNum)].Push(buffer.Bytes())
						buffer.Reset()
					}
					if err != nil {
						if err != io.EOF {
							log.Println("read log error:", err.Error())
						}
						break
					}
				}
			}(iNum)
		case <-time.After(time.Second * 3):
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
		/**
		HeapSys：程序向应用程序申请的内存
		HeapAlloc：堆上目前分配的内存
		HeapIdle：堆上目前没有使用的内存
		Alloc : 已经被配并仍在使用的字节数
		NumGC : GC次数
		HeapReleased：回收到操作系统的内存
		 */
			log.Printf("%d,%d,%d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc, m.HeapIdle, m.Alloc, m.NumGC, m.HeapReleased)
		}
	}
}
