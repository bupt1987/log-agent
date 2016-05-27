package logger

import (
	"../safe"
	"os"
	"time"
	"strconv"
	"log"
	"io"
)

type Logger interface {
	Write(n int, lQueue []*safe.Queue)
}

type FileLogger struct {
	dir string
}

func NewFileLogger(dir string) *FileLogger {
	w := FileLogger{dir}
	//创建log目录
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}
	}
	return &w
}

func (w FileLogger) Write(n int, lQueue []*safe.Queue) {
	var f *os.File
	var err error
	for {
		data := lQueue[n].Pop();
		if data == nil {
			return
		}
		file := w.dir + time.Now().Format("2006_01_02_1504") + "." + strconv.Itoa(n)
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
		d := data.([]byte)
		n, err := f.Write(d)
		if err == nil && n < len(d) {
			err = io.ErrShortWrite
		}
		if err != nil {
			log.Printf("write file %s error: %s", file, err.Error())
			continue
		}
	}
	if f != nil {
		f.Close()
	}
}
