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
	} else if url.Scheme == "tcp" {
		r, err := newTcpLogger(url.Host+url.Path, prefix)
		if err != nil {
			return nil, err
		}
		result = r.log
	}
	return result, nil
}

func (log *Logger) startCopy(pty *os.File, eofChannel chan bool, ioS *ioStruct, color string) {
	var err error
	if err = log.writePrefix(color, ioS); err != nil {
		logLib.Println(err)
	}
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
		if rune == 0x0A {
			if err = log.writePrefix(color, ioS); err != nil {
				logLib.Println(err)
			}
		}
	}
}

func (log *Logger) writePrefix(color string, ioS *ioStruct) error {
	var err error

	//Color output
	err = ioS.setTerminalColor(color)
	if err != nil {
		return err
	}

	//Write prefix
	_, err = log.output.Write([]byte(log.prefix))
	if err != nil {
		return err
	}
	//Uncolor output
	err = ioS.resetTerminal()
	if err != nil {
		return err
	}
	return nil
}
