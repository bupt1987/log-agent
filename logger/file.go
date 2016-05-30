package logger

import (
	"os"
	"time"
	"strconv"
	"log"
	"io"
)

type FileLogger struct {
	dir string
}

func NewFileLogger(dir string) FileLogger {
	w := FileLogger{dir}
	//创建log目录
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}
	}
	return w
}

func (w FileLogger) Write(pack LogPack) {
	var err error
	var f *os.File
	file := w.dir + time.Now().Format("2006_01_02_1504") + "." + strconv.Itoa(pack.num)
	f, err = os.OpenFile(file, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0664)
	if err != nil {
		log.Printf("open file %s error: %s", file, err.Error())
		return
	}
	defer f.Close()
	l, err := f.Write(pack.data)
	if err == nil && l < len(pack.data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		log.Printf("write file %s error: %s", file, err.Error())
	}
}
