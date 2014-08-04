package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
)

type ioContext struct {
	term         *terminal.Terminal
	oldTermState *terminal.State
	stdinOutput  *os.File
	log          *Logger
}

func newIOContext(stdinStr string, pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) (*ioContext, error) {
	var err error
	result := &ioContext{}

	if err = result.redirectStdin(pty, stdinStr, statusChannel); err != nil {
		return nil, errors.New("Can't redirect stdin:" + err.Error())
	}

	if err = result.redirectStdout(pty, stdout, logPrefix, logRotation, logColor, statusChannel, eofChannel); err != nil {
		return nil, errors.New("Can't redirect stdout:" + err.Error())
	}

	return result, nil
}

func (ioC *ioContext) restoreIO() error {
	err := terminal.Restore(int(ioC.stdinOutput.Fd()), ioC.oldTermState)
	return err
}

func (ioC *ioContext) redirectStdin(pty *os.File, stdinStr string, statusChannel chan ProcessStatus) error {
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
		go func() {
			io.Copy(pty, conn)
			//When the remote stdin closes, terminate the process through the status Channel
			statusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: errors.New("Remote stdin closed")}
		}()
	} else {
		//default case, the protocol is not supported
		return errors.New("The protocol " + url.Scheme + " is not supported")
	}
	ioC.stdinOutput = os.Stdin

	//Set up the tty
	ioC.oldTermState, err = terminal.MakeRaw(int(ioC.stdinOutput.Fd()))
	if err != nil {
		return errors.New("Can't create terminal:" + err.Error())
	}
	ioC.term = terminal.NewTerminal(ioC.stdinOutput, "")
	cb := func(s string, i int, r rune) (string, int, bool) {
		car := []byte{byte(r)}
		ioC.term.Write(car)
		return s, i, false
	}
	ioC.term.AutoCompleteCallback = cb

	//Copy everything from os.Stdin to the pty
	go io.Copy(pty, ioC.stdinOutput)

	return nil
}

func (ioC *ioContext) redirectStdout(pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) error {
	url, err := url.Parse(stdout)
	if err != nil {
		return err
	}
	ioC.log, err = newLogger(*url, logPrefix, logRotation, statusChannel)
	if err != nil {
		return err
	}
	go ioC.log.startCopy(pty, eofChannel, ioC, logColor)

	return nil
}

func (ioC *ioContext) setTerminalColor(color string) error {
	var err error
	switch color {
	case "red":
		_, err = ioC.term.Write(ioC.term.Escape.Red)
	case "green":
		_, err = ioC.term.Write(ioC.term.Escape.Green)
	case "blue":
		_, err = ioC.term.Write(ioC.term.Escape.Blue)
	case "yellow":
		_, err = ioC.term.Write(ioC.term.Escape.Yellow)
	case "magenta":
		_, err = ioC.term.Write(ioC.term.Escape.Magenta)
	case "cyan":
		_, err = ioC.term.Write(ioC.term.Escape.Cyan)
	case "white":
		_, err = ioC.term.Write(ioC.term.Escape.White)
	default:
		_, err = ioC.term.Write(ioC.term.Escape.Black)
	}
	return err
}

func (ioC *ioContext) resetTerminal() error {
	var err error
	_, err = ioC.term.Write(ioC.term.Escape.Reset)
	return err
}
