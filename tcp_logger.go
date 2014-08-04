package psdock

import (
	"net"
)

type tcpLogger struct {
	log *Logger
}

func newTcpLogger(path, prefix string) (*tcpLogger, error) {
	result := tcpLogger{log: &Logger{prefix: prefix}}
	conn, err := net.Dial("tcp", path)
	if err != nil {
		return nil, err
	}
	result.log.output = conn
	return &result, nil
}
