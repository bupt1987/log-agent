package logger

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type HttpLogger struct {
	to string
}

func NewHttpLogger(to string) HttpLogger {
	return HttpLogger{to}
}

func (w HttpLogger) Write(pack LogPack) {
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	var err error
	var resp *http.Response

	for n := 1; n <= 3; n ++ {
		resp, err = client.PostForm(w.to, url.Values{"log": {string(pack.data)}})
		if err != nil {
			continue
		}
		resp.Body.Close()
		if ( resp.StatusCode != http.StatusOK) {
			continue
		}
		break
	}

	if (err != nil) {
		log.Print("http post error:", err)
	} else if ( resp.StatusCode != http.StatusOK) {
		log.Printf("http error code :%d", resp.StatusCode)
	}
}

