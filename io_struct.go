package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	//"fmt"
	"io"
	"net"
	"net/url"
	"os"
)

type ioStruct struct {
	term         *terminal.Terminal
	oldTermState *terminal.State
	stdinOutput  *os.File
	log          *Logger
}

func newIOStruct(stdinStr string, pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) (*ioStruct, error) {
	var err error
	result := &ioStruct{}

	if err = result.redirectStdin(pty, stdinStr); err != nil {
		return nil, errors.New("Can't redirect stdin:" + err.Error())
	}

	if err = result.redirectStdout(pty, stdout, logPrefix, logRotation, logColor, statusChannel, eofChannel); err != nil {
		statusChannel <- ProcessStatus{Status: -1, Err: err}
	}

	return result, nil
}

func (ioS *ioStruct) restoreIO() error {
	err := terminal.Restore(int(ioS.stdinOutput.Fd()), ioS.oldTermState)
	return err
}

func (ioS *ioStruct) redirectStdin(pty *os.File, stdinStr string) error {
	url, err := url.Parse(stdinStr)
	if err != nil {
		return err
	}
	if url.Path == "os.Stdin" {
		//We don't need to do anything here
	} else if url.Scheme == "tcp" {
		conn, err := net.Dial("tcp", url.Host+url.Path)
		if err != nil {
			return err
		}
		//Directly copy from the connection to the pty. Escape chars won't be available
		go io.Copy(pty, conn)
	} else {
		//default case, the protocol is not supported
		return errors.New("The protocol " + url.Scheme + " is not supported")
	}
	ioS.stdinOutput = os.Stdin

	//Set up the tty
	ioS.oldTermState, err = terminal.MakeRaw(int(ioS.stdinOutput.Fd()))
	if err != nil {
		return errors.New("Can't create terminal:" + err.Error())
	}
	ioS.term = terminal.NewTerminal(ioS.stdinOutput, "")
	cb := func(s string, i int, r rune) (string, int, bool) {
		car := []byte{byte(r)}
		ioS.term.Write(car)
		return s, i, false
	}
	ioS.term.AutoCompleteCallback = cb

	//Copy everything from os.Stdin to the pty
	go io.Copy(pty, ioS.stdinOutput)

	return nil
}

func (ioS *ioStruct) redirectStdout(pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) error {
	url, err := url.Parse(stdout)
	if err != nil {
		return err
	}
	ioS.log, err = newLogger(*url, logPrefix, logRotation, statusChannel)
	if err != nil {
		return err
	}
	go ioS.log.startCopy(pty, eofChannel, ioS, logColor)

	return nil
}

func (ioS *ioStruct) setTerminalColor(color string) error {
	var err error
	switch color {
	case "red":
		_, err = ioS.term.Write(ioS.term.Escape.Red)
	case "green":
		_, err = ioS.term.Write(ioS.term.Escape.Green)
	case "blue":
		_, err = ioS.term.Write(ioS.term.Escape.Blue)
	case "yellow":
		_, err = ioS.term.Write(ioS.term.Escape.Yellow)
	case "magenta":
		_, err = ioS.term.Write(ioS.term.Escape.Magenta)
	case "cyan":
		_, err = ioS.term.Write(ioS.term.Escape.Cyan)
	case "white":
		_, err = ioS.term.Write(ioS.term.Escape.White)
	default:
		_, err = ioS.term.Write(ioS.term.Escape.Black)
	}
	return err
}

func (ioS *ioStruct) resetTerminal() error {
	var err error
	_, err = ioS.term.Write(ioS.term.Escape.Reset)
	return err
}
