package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"time"
)

func (p *Process) redirectStdin() error {
	var err error
	p.oldTermState, err = terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return errors.New("Can't redirect stdin:" + err.Error())
	}
	newTerminal := terminal.NewTerminal(os.Stdin, "")
	cb := func(s string, i int, r rune) (string, int, bool) {
		car := []byte{byte(r)}
		newTerminal.Write(car)
		return s, i, false
	}
	newTerminal.AutoCompleteCallback = cb
	go io.Copy(p.Pty, os.Stdin)
	return nil
}

func (p *Process) restoreStdin() error {
	err := terminal.Restore(int(os.Stdin.Fd()), p.oldTermState)
	return err
}

func (p *Process) redirectStdout() error {
	if p.Conf.Stdout != "os.Stdout" {
		url, err := url.Parse(p.Conf.Stdout)
		if err != nil {
			return err
		}
		if url.Scheme == "file" {
			go p.manageLogRotation(url.Host + url.Path)
			//Wait for the file to be ready
			time.Sleep(100 * time.Millisecond)
		} else if url.Scheme == "tcp" {
			p.output, err = net.Dial("tcp", url.Host+url.Path)
			if err != nil {
				p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
			}
		}
	}
	log.SetOutput(p.output)
	log.SetPrefix(p.Conf.LogPrefix)
	go io.Copy(p.output, p.Pty)
	return nil
}

func (p *Process) manageLogRotation(filename string) {
	var lifetime time.Duration
	switch p.Conf.LogRotation {
	case "minutely":
		lifetime = time.Minute
	case "hourly":
		lifetime = time.Hour
	case "daily":
		lifetime = time.Hour * 24
	case "weekly":
		lifetime = time.Hour * 24 * 7
	}
	for {
		f, err := os.Create(filename + "-" + string(time.Now().Format("2006-01-02-15-04")))
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		p.output = f
		time.Sleep(lifetime)
	}
}
