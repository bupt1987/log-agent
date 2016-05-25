package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"io"
	"bytes"
	"fmt"
	"time"
	"log"
)

type logPack struct {
	key     string
	content string
}

var SOCKET string = "/tmp/go-unix.socket"
var LOGDIR string = "/tmp/go-unix-log/"
var LOGMAX int = 10000;

var LOGKEY string = ""
var LOGBUFFER bytes.Buffer

func flashLog() {
	if LOGKEY == "" {
		return
	}

	content := LOGBUFFER.String()

	if (content == "") {
		fmt.Println("buffer is empty")
		return
	}

	var f *os.File
	var err error

	file := LOGDIR + LOGKEY + ".log"

	if _, err = os.Stat(file); err == nil {
		f, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY, os.ModeAppend)
	} else {
		f, err = os.Create(file)
	}

	if err != nil {
		log.Printf("open file %s error: %s", file, err.Error())
		return
	}
	//写文件
	writer := bufio.NewWriter(f)
	defer f.Close()
	if _, err = fmt.Fprint(writer, content); err != nil {
		log.Printf("write file %s error: %s", file, err.Error())
		return
	}

	if err = writer.Flush(); err != nil {
		log.Printf("flush file %s error: %s", file, err.Error())
		return
	}

	fmt.Printf("save file %s\n", file)
	LOGBUFFER.Reset()
}

func main() {
	lognum := 0;

	connchan := make(chan net.Conn, 128)
	logchan := make(chan logPack, 128)

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, os.Kill)
	signal.Notify(sigchan, syscall.SIGTERM)

	//删除socket文件
	if _, err := os.Stat(SOCKET); err == nil {
		if err := os.Remove(SOCKET); err != nil {
			panic(err)
		}
	}

	//创建log目录
	if _, err := os.Stat(LOGDIR); err != nil {
		if err := os.MkdirAll(LOGDIR, 0777); err != nil {
			panic(err)
		}
	}

	//监听
	// "unix", "unixgram" and "unixpacket".
	linsten, err := net.Listen("unix", SOCKET)
	if err != nil {
		panic(err)
	}

	if err := os.Chmod(SOCKET, 0777); err != nil {
		panic(err)
	}

	defer linsten.Close()

	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Print("get client connection error: ", err)
				continue
			}
			connchan <- conn
		}
	}(linsten)

	for {
		select {
		case <-sigchan:
			log.Println("exit")
			flashLog()
			if err := os.Remove(SOCKET); err != nil {
				panic(err)
			}
			return
		case conn := <-connchan:
			go func(conn net.Conn) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				var buffer bytes.Buffer
				for {
					data, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							if data != "" {
								buffer.WriteString(data)
							}
							break
						}
						fmt.Println(err)
						return
					}
					buffer.WriteString(data)
				}

				logchan <- logPack{time.Now().Format("2006_01_02_1504"), buffer.String()}

			}(conn)
		case log := <-logchan:
			if LOGKEY == "" {
				LOGKEY = log.key
			}

			if LOGKEY != log.key || lognum >= LOGMAX {
				flashLog()
				LOGKEY = log.key
				lognum = 0;
			}
			if (log.content != "") {
				LOGBUFFER.WriteString(log.content)
				lognum ++
				//fmt.Printf("[%d]add new log %s : %#v\n", lognum, log.key, log.content)
			}
		case <-time.After(time.Second * 60):
			fmt.Println("auto flush log")
			flashLog()
		}
	}
}
