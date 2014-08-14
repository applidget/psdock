package psdock

import "crypto/tls"

//tlsLogger is the type used to communicate through a TCP connection
type tlsLogger struct {
	log *Logger
}

func newTlsLogger(path, prefix string) (*tlsLogger, error) {
	result := tlsLogger{log: &Logger{prefix: prefix}}
	conf := &tls.Config{InsecureSkipVerify: false}
	conn, err := tls.Dial("tcp", path, conf)
	if err != nil {
		return nil, err
	}
	result.log.output = conn
	return &result, nil
}
