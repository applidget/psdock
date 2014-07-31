package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"io"
	"os"
)

type stdin struct {
	term         *terminal.Terminal
	oldTermState *terminal.State
	stdinOutput  *os.File
}

func setTerminalAndRedirectStdin(newStdin, pty *os.File) (*stdin, error) {
	var err error
	result := &stdin{stdinOutput: newStdin}
	result.oldTermState, err = terminal.MakeRaw(int(newStdin.Fd()))
	if err != nil {
		return nil, errors.New("Can't redirect stdin:" + err.Error())
	}
	result.term = terminal.NewTerminal(newStdin, "")
	cb := func(s string, i int, r rune) (string, int, bool) {
		car := []byte{byte(r)}
		result.term.Write(car)
		return s, i, false
	}
	result.term.AutoCompleteCallback = cb

	go io.Copy(pty, newStdin)

	return result, nil
}

func (s *stdin) restoreStdin() error {
	err := terminal.Restore(int(s.stdinOutput.Fd()), s.oldTermState)
	return err
}

func (s *stdin) setTerminalColor(color string) error {
	var err error
	switch color {
	case "red":
		_, err = s.term.Write(s.term.Escape.Red)
	case "green":
		_, err = s.term.Write(s.term.Escape.Green)
	case "blue":
		_, err = s.term.Write(s.term.Escape.Blue)
	case "yellow":
		_, err = s.term.Write(s.term.Escape.Yellow)
	case "magenta":
		_, err = s.term.Write(s.term.Escape.Magenta)
	case "cyan":
		_, err = s.term.Write(s.term.Escape.Cyan)
	case "white":
		_, err = s.term.Write(s.term.Escape.White)
	default:
		_, err = s.term.Write(s.term.Escape.Black)
	}
	return err
}

func (s *stdin) resetTerminal() error {
	var err error
	_, err = s.term.Write(s.term.Escape.Reset)
	return err
}
