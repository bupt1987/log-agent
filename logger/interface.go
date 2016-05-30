package logger

import "log"

type LogPack struct {
	num  int
	data []byte
}

type Logger interface {
	Write(pack LogPack)
}

func NewPack(n int, data []byte) LogPack {
	return LogPack{n, data}
}

func NewLogger(t string, to string) Logger {
	switch t {
	case "file":
		return NewFileLogger(to)
	case "http":
		return NewHttpLogger(to)
	default:
		log.Fatal("input -type error, must be file or http")
	}
	return nil
}
