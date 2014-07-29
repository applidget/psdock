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

func redirectStdin(newStdin, pty *os.File) (*stdin, error) {
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
