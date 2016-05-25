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
	"io/ioutil"
	"io"
)

type logPack struct {
	key     string
	content []byte
}

var SOCKET string = "/tmp/go-unix.socket"
var LOGDIR string = "/tmp/go-unix-log/"
var LOGMAX int = 1000;

var LOGKEY string = ""
var LOGBUFFER bytes.Buffer

func flashLog() {
	if LOGKEY == "" {
		return
	}

	data := LOGBUFFER.Bytes()
	length := len(data)

	if (length == 0) {
		log.Println("buffer is empty")
		return
	}

	var f *os.File
	var err error

	file := LOGDIR + "debug.log"

	f, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0644)

	if err != nil {
		log.Printf("open file %s error: %s", file, err.Error())
		return
	}
	//写文件
	n, err := f.Write(data)
	if err == nil && n < length {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		log.Printf("write file %s error: %s", file, err.Error())
		return
	}
	log.Printf("save file %s\n", file)
	LOGBUFFER.Reset()
}

func main() {
	lognum := 0;

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
				log.Println("connection error: ", err)
				continue
			}
			select {
			case connchan <- conn:
			//do nothing
			default:
			//warnning!
				log.Println("CONN_CHANNEL is full!")
			}

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
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Println(err)
					return
				}
				select {
				case logchan <- logPack{time.Now().Format("2006_01_02_1504"), data}:
				//do nothing
				default:
				//warnning!
					log.Println("LOG_CHANNEL is full!")
				}
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
			if (len(log.content) != 0) {
				LOGBUFFER.Write(log.content)
				lognum ++
				//log.Printf("[%d]add new log %s : %#v\n", lognum, log.key, log.content)
			}
		case <-time.After(time.Second * 60):
			log.Println("auto flush log")
			flashLog()
		}
	}
}
