package psdock

import (
	"bufio"
	"errors"
	"io"
	logLib "log"
	"net/url"
	"os"
	"strings"
)

type Logger struct {
	output          io.WriteCloser
	prefix          string
	writePrefixNext bool
}

func newLogger(url url.URL, prefix string, lRotation string, statusChannel chan ProcessStatus) (*Logger, error) {
	var result *Logger
	if url.Path == "os.Stdout" {
		result = &Logger{output: os.Stdout, prefix: prefix}
	} else if url.Scheme == "file" {
		var err error
		r, err := NewFileLogger(url.Host+url.Path, prefix, lRotation, statusChannel)
		if err != nil {
			return nil, errors.New("Error in newLogger" + err.Error())
		}
		err = r.openFirstOutputFile()
		if err != nil {
			return nil, errors.New("Error in newLogger" + err.Error())
		}
		result = r.log
	} else if url.Scheme == "tcp" {
		r, err := newTcpLogger(url.Host+url.Path, prefix)
		if err != nil {
			return nil, errors.New("Error in newLogger" + err.Error())
		}
		result = r.log
	} else if url.Scheme == "tls" {
		//If it is a TLS Logger, split the url : the first part is used to connect to the socket, the second is written on it
		r, err := newTlsLogger(url.Host, prefix)
		if err != nil {
			return nil, errors.New("Error in newLogger" + err.Error())
		}
		result = r.log
		header := strings.Join([]string{url.Path[1 : len(url.Path)-2], "%0A%0A"}, "")
		_, err = result.output.Write([]byte(header))
		if err != nil {
			return nil, errors.New("Error in newLogger" + err.Error())
		}
	} else {
		//default case, the protocol is not supported
		return nil, errors.New("The protocol " + url.Scheme + " is not supported")
	}
	return result, nil
}

//startCopy copies the stdout from pty to the file pointed at by log.output.
//When a reading error is raised, or the EOF symbol is read, the eofChannel is trigerred
func (log *Logger) startCopy(pty *os.File, eofChannel chan bool, ioC *ioContext, color string) {
	var err error
	if err = log.writePrefix(color, ioC); err != nil {
		logLib.Println("Error in Logger.startCopy():" + err.Error())
	}
	reader := bufio.NewReader(pty)
	for {
		rune, _, err := reader.ReadRune()
		if err == io.EOF {
			eofChannel <- true
			return
		}
		if err != nil {
			if err.Error() == "read /dev/ptmx: input/output error" {
				//hack to make it work in a container
				eofChannel <- true
				return
			}
			logLib.Println("Error in Logger.startCopy():Error in reader.ReadRune():" + err.Error())
			break
		}
		if log.writePrefixNext {
			if err = log.writePrefix(color, ioC); err != nil {
				logLib.Println("Error in Logger.startCopy():" + err.Error())
			}
			log.writePrefixNext = false
		}
		_, err = log.output.Write([]byte{byte(rune)})
		if err != nil {
			logLib.Println("Error in Logger.startCopy():Error in Write():" + err.Error())
			break
		}
		if rune == EOL { //If we just read an end-of-line, write the prefix
			log.writePrefixNext = true
		}
	}
}

//writePrefix writes the prefix on the stdout
func (log *Logger) writePrefix(color string, ioC *ioContext) error {
	var err error

	//Color output
	err = ioC.setTerminalColor(color)
	if err != nil {
		return errors.New("Error in Logger.writeprefix():" + err.Error())
	}

	//Write prefix
	_, err = log.output.Write([]byte(log.prefix))
	if err != nil {
		return errors.New("Error in Logger.writeprefix():" + err.Error())
	}
	//Uncolor output
	err = ioC.resetTerminal()
	if err != nil {
		return errors.New("Error in Logger.writeprefix():" + err.Error())
	}
	return nil
}
