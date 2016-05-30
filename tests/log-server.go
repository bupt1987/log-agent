package main

import (
	"net/http"
	"log"
	"io"
	"time"
	"sync/atomic"
)

var totalNum int32 = 0

func HelloServer(w http.ResponseWriter, req *http.Request) {
	//log.Print("log:", req.PostFormValue("log"))
	atomic.AddInt32(&totalNum, 1)
	io.WriteString(w, "OK\n")
}

func main() {
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for _ = range ticker.C {
			log.Printf("request: %v\n", totalNum)
		}
	}()

	http.HandleFunc("/upload", HelloServer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
