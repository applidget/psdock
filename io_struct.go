package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"io"
	"log"
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

	//Since psdock is to be used in lxc containers, it won't be able to redirect the stdin in that case. Therefore, we do not fail in that case,
	//but only signal the error
	if err = result.redirectStdin(pty, stdinStr, statusChannel); err != nil {
		log.Println("Error in newIOContext:" + err.Error())
	}
	if err = result.redirectStdout(pty, stdout, logPrefix, logRotation, logColor, statusChannel, eofChannel); err != nil {
		return nil, errors.New("Error in newIOContext:" + err.Error())
	}

	return result, nil
}

func (ioC *ioContext) restoreIO() error {
	if ioC.oldTermState == nil {
		return nil
	}
	err := terminal.Restore(int(ioC.stdinOutput.Fd()), ioC.oldTermState)
	ioC.oldTermState = nil
	return err
}

//redirectStdin parses stdinStr to determine the input, then setups the tty, starts the copy and returns
func (ioC *ioContext) redirectStdin(pty *os.File, stdinStr string, statusChannel chan ProcessStatus) error {
	url, err := url.Parse(stdinStr)
	if err != nil {
		return errors.New("Error in ioContext.redirectStdin():" + err.Error())
	}
	if url.Path == "os.Stdin" {
		//We don't need to do anything here
	} else if url.Scheme == "tcp" {
		conn, err := net.Dial("tcp", url.Host+url.Path)
		if err != nil {
			return errors.New("Error in ioContext.redirectStdin():" + err.Error())
		}
		//Directly copy from the connection to the pty. Escape chars won't be available
		go func() {
			io.Copy(pty, conn)
			//When the remote stdin closes, terminate the process through the status Channel
			statusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: errors.New("Error in ioContext.redirectStdin():Remote stdin closed")}
		}()
	} else {
		//default case, the protocol is not supported
		return errors.New("Error in ioContext.redirectStdin():The protocol " + url.Scheme + " is not supported")
	}
	ioC.stdinOutput = os.Stdin

	//Set up the tty. We make sure that we do not fail BEFORE having created a term object
	ioC.oldTermState, err = terminal.MakeRaw(int(ioC.stdinOutput.Fd()))
	ioC.term = terminal.NewTerminal(ioC.stdinOutput, "")
	if err != nil {
		return errors.New("Error in ioContext.redirectStdin():Can't create terminal:" + err.Error())
	}

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

//redirectStdout parses stdout, creates a logger for it, starts the copy and returns
func (ioC *ioContext) redirectStdout(pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) error {
	url, err := url.Parse(stdout)
	if err != nil {
		return errors.New("Error in ioContext.redirectStdout:" + err.Error())
	}
	ioC.log, err = newLogger(*url, logPrefix, logRotation, statusChannel)
	if err != nil {
		return errors.New("Error in ioContext.redirectStdout:" + err.Error())
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
		_, err = ioC.term.Write(ioC.term.Escape.White)
	}
	return err
}

func (ioC *ioContext) resetTerminal() error {
	var err error
	_, err = ioC.term.Write(ioC.term.Escape.Reset)
	return err
}
