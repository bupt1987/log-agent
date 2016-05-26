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

var sSocket string = "/tmp/go-unix.socket"
var sLogDir string = "/go/logs/go-unix/"
var sLogKey string = ""
var bufLog bytes.Buffer
var oLog *os.File
var iLogMaxSize int = 1024 * 1024 * 1

func getKey() string {
	return time.Now().Format("2006_01_02_1504")
}

func flashLog() {
	if sLogKey == "" {
		return
	}
	data := bufLog.Bytes()
	length := len(data)

	if (length == 0) {
		return
	}
	var err error
	file := sLogDir + sLogKey + ".log"

	if oLog == nil {
		oLog, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0664)
	}

	if err != nil {
		log.Printf("open file %s error: %s", file, err.Error())
		return
	}
	//写文件
	n, err := oLog.Write(data)
	if err == nil && n < length {
		err = io.ErrShortWrite
	}
	if err != nil {
		log.Printf("write file %s error: %s", file, err.Error())
		return
	}
	log.Printf("save file %s\n", file)
	bufLog.Reset()
}

func main() {
	chConn := make(chan net.Conn, 1024)
	chLog := make(chan logPack, 1024)

	chSig := make(chan os.Signal)
	signal.Notify(chSig, os.Interrupt)
	signal.Notify(chSig, os.Kill)
	signal.Notify(chSig, syscall.SIGTERM)

	//删除socket文件
	if _, err := os.Stat(sSocket); err == nil {
		if err := os.Remove(sSocket); err != nil {
			panic(err)
		}
	}

	//创建log目录
	if _, err := os.Stat(sLogDir); err != nil {
		if err := os.MkdirAll(sLogDir, 0777); err != nil {
			panic(err)
		}
	}

	//监听
	linsten, err := net.Listen("unix", sSocket)
	if err != nil {
		panic(err)
	}

	if err := os.Chmod(sSocket, 0777); err != nil {
		panic(err)
	}

	defer linsten.Close()

	go func() {
		for {
			conn, err := linsten.Accept()
			if err != nil {
				log.Println("connection error:", err)
				continue
			}
			chConn <- conn
		}
	}()

	go func() {
		for {
			select {
			case data := <-chLog:
				if sLogKey == "" {
					sLogKey = data.key
				}
				if sLogKey != data.key {
					flashLog()
					sLogKey = data.key
					oLog.Close()
					oLog = nil
				} else if bufLog.Len() >= iLogMaxSize {
					flashLog()
				}
				bufLog.Write(data.content)
			case <-time.After(time.Second * 60):
				key := getKey()
				if sLogKey == "" || sLogKey == key {
					continue
				}
				log.Println("auto flush log", sLogKey)
				flashLog()
				sLogKey = key
				if oLog != nil {
					oLog.Close()
					oLog = nil
				}
			}
		}
	}()

	for {
		select {
		case <-chSig:
			if _, err := os.Stat(sSocket); err == nil {
				if err := os.Remove(sSocket); err != nil {
					panic(err)
				}
			}
			flashLog()
			return
		case conn := <-chConn:
			go func() {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				for {
					data, err := reader.ReadBytes('\n')
					if err != nil {
						if err != io.EOF {
							log.Println("read data error: ", err)
							break
						}
						if len(data) > 0 {
							chLog <- logPack{getKey(), data}
						}
						break
					}
					chLog <- logPack{getKey(), data}
				}
			}()
		}
	}
}
