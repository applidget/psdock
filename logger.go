package psdock

import (
	"bufio"
	"io"
	logLib "log"
	"net/url"
	"os"
)

type Logger struct {
	output io.WriteCloser
	prefix string
}

func newLogger(url url.URL, prefix string, lRotation string, statusChannel chan ProcessStatus) (*Logger, error) {
	var result *Logger
	if url.Path == "os.Stdout" {
		result = &Logger{output: os.Stdout, prefix: prefix}
	}
	if url.Scheme == "file" {
		var err error
		r, err := NewFileLogger(url.Host+url.Path, prefix, lRotation, statusChannel)
		if err != nil {
			return nil, err
		}
		err = r.openFirstOutputFile()
		if err != nil {
			return nil, err
		}
		result = r.log
	} /*else if url.Scheme == "tcp" {
		result, err = newTcpLogger("tcp", url.Host+url.Path, prefix)
		//result, err = net.Dial("tcp", url.Host+url.Path)
		if err != nil {
			return nil, err
		}
	}*/
	return result, nil
}

func (log *Logger) startCopy(pty *os.File, eofChannel chan bool) {
	_, _ = log.output.Write([]byte(log.prefix))
	reader := bufio.NewReader(pty)
	for {
		rune, _, err := reader.ReadRune()
		if err == io.EOF {
			eofChannel <- true
			return
		}
		if err != nil {
			logLib.Println("erreur")
			logLib.Println(err)
			break
		}
		_, err = log.output.Write([]byte{byte(rune)})
		if err != nil {
			logLib.Println("erreur")
			logLib.Println(err)
			break
		}
	}
}
