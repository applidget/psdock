package psdock

import (
	"bufio"
	"io"
	"net/url"
	"os"
)

type Logger struct {
	output io.WriteCloser
	prefix string
}

func newLogger(url url.URL, prefix string, lRotation string) (*Logger, error) {
	var result Logger
	if url.Scheme == "file" {
		var err error
		r, err := NewFileLogger(url.Host+url.Path, prefix, lRotation)
		result = r.log
		err = r.openFirstOutputFile()
		if err != nil {
			return nil, err
		}
	} /*else if url.Scheme == "tcp" {
		result, err = newTcpLogger("tcp", url.Host+url.Path, prefix)
		//result, err = net.Dial("tcp", url.Host+url.Path)
		if err != nil {
			return nil, err
		}
	}*/
	return &result, nil
}

func (log *Logger) startCopy(pty *os.File) {
	_, _ = log.output.Write([]byte(log.prefix))
	reader := bufio.NewReader(pty)
	for {
		rune, _, _ := reader.ReadRune()
		_, err := log.output.Write([]byte{byte(rune)})
		if rune == '\n' {
			_, _ = log.output.Write([]byte(log.prefix))
		}
		if err != nil {
			break
		}
	}
	//If we arrive here, the logger has created a new file, and it is assigned to p.output
	//We start writing on the new p.output
	log.startCopy(pty)
}
