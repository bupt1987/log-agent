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
)

func main() {
	socket := *flag.String("socket", "/tmp/log-agent.socket", "Listen socket address")
	dir := *flag.String("to", "/go/logs/", "Log store to address")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [name ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	iNum := runtime.NumCPU()
	lQueue := make([]*safe.Queue, iNum)
	chConn := make(chan net.Conn, 10)
	chSig := make(chan os.Signal)
	signal.Notify(chSig, os.Interrupt)
	signal.Notify(chSig, os.Kill)
	signal.Notify(chSig, syscall.SIGTERM)

	log.Println(socket)

	//删除socket文件
	if _, err := os.Stat(socket); err == nil {
		if err := os.Remove(socket); err != nil {
			panic(err)
		}
	}

	//监听
	listen, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}

	if err := os.Chmod(socket, 0777); err != nil {
		panic(err)
	}

	for i := 0; i < iNum; i ++ {
		lQueue[i] = safe.NewQueue()
	}

	for n := 0; n < iNum; n ++ {
		go func(n int, lQueue []*safe.Queue) {
			writer := logger.NewFileLogger(dir)
			for {
				writer.Write(n, lQueue)
				time.Sleep(time.Second)
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

	log.Println("running")

	for {
		select {
		case <-chSig:
			if _, err := os.Stat(socket); err == nil {
				if err := os.Remove(socket); err != nil {
					panic(err)
				}
			}
			return
		case conn := <-chConn:
			go func(iNum int) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				var buffer bytes.Buffer
				for {
					data, err := reader.ReadBytes('\n')
					if len(data) > 0 {
						buffer.Write(data)
					}
					if (buffer.Len() >= 512000 || err == io.EOF) {
						lQueue[int(time.Now().Unix()) % iNum].Push(buffer.Bytes())
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
		}

	}
}
