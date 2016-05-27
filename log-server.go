package main
/**
dev
2016/05/27 01:39:36 total:  25600000
2016/05/27 01:39:36 error:  0
2016/05/27 01:39:36 run time: 32697.62 ms

2016/05/27 14:15:31 total:  25600000
2016/05/27 14:15:31 error:  0
2016/05/27 14:15:31 run time: 23124.78 ms

2016/05/27 14:17:42 total:  25600000
2016/05/27 14:17:42 error:  0
2016/05/27 14:17:42 run time: 10096.48 ms

master
2016/05/27 01:43:56 total:  25600000
2016/05/27 01:43:56 error:  0
2016/05/27 01:43:56 run time: 49095.38 ms
 */
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
	"./writer"
	"time"
)

func main() {
	var sSocket string = "/tmp/go-unix.socket"
	var sLogDir string = "/go/logs/go-unix/"

	var iCpu int = runtime.NumCPU()
	lQueue := make([]*safe.Queue, iCpu)
	chConn := make(chan net.Conn, 1024)
	chSig := make(chan os.Signal)
	logWriter := writer.NewWriter(sLogDir)

	signal.Notify(chSig, os.Interrupt)
	signal.Notify(chSig, os.Kill)
	signal.Notify(chSig, syscall.SIGTERM)

	for i := 0; i < iCpu; i ++ {
		lQueue[i] = safe.NewQueue()
	}

	//删除socket文件
	if _, err := os.Stat(sSocket); err == nil {
		if err := os.Remove(sSocket); err != nil {
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

	for n := 0; n < iCpu; n ++ {
		go func(n int, lQueue []*safe.Queue) {
			for {
				logWriter.Write(n, lQueue)
				time.Sleep(time.Millisecond * 3)
			}
		}(n, lQueue)
	}

	log.Println("running")

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
			go func(iCpu int) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				var buffer bytes.Buffer
				for {
					data, err := reader.ReadBytes('\n')
					if err == nil {
						buffer.Write(data)
					}
					if (buffer.Len() >= 512000 || err == io.EOF) {
						lQueue[int(time.Now().Unix()) % iCpu].Push(buffer.Bytes())
						buffer.Reset()
					}
					if err != nil {
						if err != io.EOF {
							log.Println("read log error:", err.Error())
						}
						break
					}
				}
			}(iCpu)
		}
	}
}
