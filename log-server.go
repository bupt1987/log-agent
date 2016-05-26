package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"time"
	"log"
	"io"
	"test/safe"
	"runtime"
	"strconv"
)

var sSocket string = "/tmp/go-unix.socket"
var sLogDir string = "/go/logs/go-unix/"

func main() {
	iCpu := runtime.NumCPU()
	queue := safe.NewQueue()
	chConn := make(chan net.Conn, 1024)

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

	go func(linsten net.Listener) {
		for {
			conn, err := linsten.Accept()
			if err != nil {
				log.Println("connection error:", err)
				continue
			}
			chConn <- conn
		}
	}(linsten)

	for n := 1; n <= iCpu; n ++ {
		go func(n int) {
			var f *os.File
			for {
				data := queue.Pop();
				if data != nil {
					//写文件
					file := sLogDir + time.Now().Format("2006_01_02_1504") + "." + strconv.Itoa(n)
					if f == nil || file != f.Name() {
						if f != nil {
							f.Close()
						}
						f, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0664)
						if err != nil {
							log.Printf("open file %s error: %s", file, err.Error())
							continue
						}
					}
					//写文件
					b := data.([]byte)
					n, err := f.Write(b)
					if err == nil && n < len(b) {
						err = io.ErrShortWrite
					}
					if err == os.ErrNotExist {
						f, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0664)
						if err != nil {
							log.Printf("open file %s error: %s", file, err.Error())
							continue
						}
					}
					if err != nil {
						log.Printf("write file %s error: %s", file, err.Error())
						continue
					}
				} else {
					//select {
					//case <-time.After(time.Second * 60):
					//	if f == nil {
					//		continue
					//	}
					//	file := sLogDir + time.Now().Format("2006_01_02_1504") + "." + strconv.Itoa(n)
					//	if f.Name() != file {
					//		f.Close()
					//	}
					//}
					time.Sleep(time.Millisecond * 1)
				}
			}
		}(n)
	}

	for {
		select {
		case <-chSig:
			if _, err := os.Stat(sSocket); err == nil {
				if err := os.Remove(sSocket); err != nil {
					panic(err)
				}
			}
			return
		case conn := <-chConn:
			go func() {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				for {
					data, err := reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Println("read data error: ", err)
						break
					}
					queue.Push(data)
				}
			}()
		}
	}
}
