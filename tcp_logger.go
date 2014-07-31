package psdock

import (
	"net"
)

type tcpLogger struct {
	log *Logger
}

func newTcpLogger(protocol, path, prefix string) (*tcpLogger, error) {
	result := tcpLogger{log: &Logger{prefix: prefix}}
	conn, err := net.Dial(protocol, path)
	if err != nil {
		return nil, err
	}
	result.log.output = conn
	return &result, nil
}
