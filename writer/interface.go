package writer

import "../safe"

type LogWriter interface {
	Write(n int, lQueue []*safe.Queue)
}
