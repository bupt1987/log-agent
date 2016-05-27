package writer

import (
	"../safe"
	"os"
	"time"
	"strconv"
	"log"
	"io"
)

type FileWriter struct {
	dir string
}

func NewWriter(dir string) *FileWriter {
	w := FileWriter{dir}
	//创建log目录
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}
	}
	return &w
}

func (w FileWriter) Write(n int, lQueue []*safe.Queue) {
	var f *os.File
	var err error
	for {
		data := lQueue[n].Pop();
		if data == nil {
			return
		}
		//写文件
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
	}
	if f != nil {
		f.Close()
	}
}
