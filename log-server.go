package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"bytes"
	"time"
	"log"
	"io"
)

type logPack struct {
	key     string
	content []byte
}

var SOCKET string = "/tmp/go-unix.socket"
var LOGDIR string = "/go/logs/go-unix/"

var LOGKEY string = ""
var LOGBUFFER bytes.Buffer
var LOGHANDLER *os.File
var LOGMAXSIZE int = 1024 * 1024 * 10

func getKey() string {
	return time.Now().Format("2006_01_02_1504")
}

func flashLog() {
	if LOGKEY == "" {
		return
	}

	data := LOGBUFFER.Bytes()
	length := len(data)

	if (length == 0) {
		return
	}

	var err error

	file := LOGDIR + LOGKEY + ".log"

	if LOGHANDLER == nil {
		LOGHANDLER, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0664)
	}

	if err != nil {
		log.Printf("open file %s error: %s", file, err.Error())
		return
	}
	//写文件
	n, err := LOGHANDLER.Write(data)
	if err == nil && n < length {
		err = io.ErrShortWrite
	}
	if err != nil {
		log.Printf("write file %s error: %s", file, err.Error())
		return
	}
	log.Printf("save file %s\n", file)
	LOGBUFFER.Reset()
}

func pushLog(logchan chan logPack, data []byte, force bool) bool {

	if (len(data) == 0) {
		return false
	}

	select {
	case logchan <- logPack{getKey(), data}:
		return true
	}

	if force {
		for {
			select {
			case logchan <- logPack{getKey(), data}:
				return true
			}
		}
	}
	return false
}

func main() {
	connchan := make(chan net.Conn, 1024)
	logchan := make(chan logPack, 1024)

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
				log.Println("connection error: ", err)
				continue
			}
			select {
			case connchan <- conn:
			default:
				conn.Close()
				log.Println("connection channel is full!")
			}
		}
	}(linsten)

	for {
		select {
		case <-sigchan:
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
					data, err := reader.ReadBytes('\n')
					if err != nil {
						if err != io.EOF {
							log.Println("read data error: ", err)
							break
						}
						buffer.Write(data)
						pushLog(logchan, buffer.Bytes(), true)
						break
					}
					buffer.Write(data)
					bPush := pushLog(logchan, buffer.Bytes(), false)
					if bPush {
						buffer.Reset()
					}
				}
			}(conn)
		case data := <-logchan:
			if LOGKEY == "" {
				LOGKEY = data.key
			}
			if LOGKEY != data.key {
				flashLog()
				LOGKEY = data.key
				LOGHANDLER.Close()
				LOGHANDLER = nil
			} else if LOGBUFFER.Len() >= LOGMAXSIZE {
				flashLog()
			}
			LOGBUFFER.Write(data.content)
			time.Sleep(time.Second * 10)
		//log.Printf("new log %s : %s", data.key, data.content)
		case <-time.After(time.Second * 60):
			key := getKey()
			if LOGKEY == "" || LOGKEY == key {
				continue
			}
			log.Println("auto flush log ", LOGKEY)
			flashLog()
			LOGKEY = key
			if LOGHANDLER != nil {
				LOGHANDLER.Close()
				LOGHANDLER = nil
			}
		}
	}
}
