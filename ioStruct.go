package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"io"
	"net/url"
	"os"
)

type ioStruct struct {
	term         *terminal.Terminal
	oldTermState *terminal.State
	stdinOutput  *os.File
	log          *Logger
}

func newIOStruct(newStdin, pty *os.File, stdout, logPrefix, logRotation, logColor string,
	statusChannel chan ProcessStatus, eofChannel chan bool) (*ioStruct, error) {
	var err error
	result := &ioStruct{stdinOutput: newStdin}
	result.oldTermState, err = terminal.MakeRaw(int(newStdin.Fd()))
	if err != nil {
		return nil, errors.New("Can't create terminal:" + err.Error())
	}
	result.term = terminal.NewTerminal(newStdin, "")
	cb := func(s string, i int, r rune) (string, int, bool) {
		car := []byte{byte(r)}
		result.term.Write(car)
		return s, i, false
	}
	result.term.AutoCompleteCallback = cb

	go io.Copy(pty, newStdin)

	if err = result.redirectStdout(pty, stdout, logPrefix, logRotation, logColor, statusChannel, eofChannel); err != nil {
		statusChannel <- ProcessStatus{Status: -1, Err: err}
	}

	return result, nil
}

func (ioS *ioStruct) restoreIO() error {
	err := terminal.Restore(int(ioS.stdinOutput.Fd()), ioS.oldTermState)
	return err
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
