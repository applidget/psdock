package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
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
	go p.startCopy()
	return nil
}

func (p *Process) startCopy() {
	_, _ = io.Copy(p.output, p.Pty)
	//If we arrive here, the logger has created a new file.
	//We start writing on the new p.Output
	p.startCopy()
}

func (p *Process) manageLogRotation(filename string) {
	var previousName, newName string
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
		newName = filename + "-" + string(time.Now().Format("2006-01-02-15-04"))
		f, err := os.Create(newName)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		oldOutput := p.output
		p.output = f
		if previousName != "" {
			//We close the old file so that startCopy has to call again io.Copy to the updated p.output
			if err = oldOutput.Close(); err != nil {
				log.Print(err)
			}
			//previousName is ready to be gzipped
			if err := compressOldOutput(previousName); err != nil {
				log.Print("Can't archive old file:" + err.Error())
			}
		}
		previousName = newName
		time.Sleep(lifetime)
	}
}

//compressOldOutput creates a gzip archive oldFile.gz and puts oldFilePath in it
//it then removes oldFile
func compressOldOutput(oldFile string) error {
	file, err := os.Create(oldFile + ".gz")
	if err != nil {
		return err
	}
	defer file.Close()
	gWriter := gzip.NewWriter(file)
	defer gWriter.Close()
	fileContent, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return err
	}
	if _, err = gWriter.Write(fileContent); err != nil {
		return err
	}
	if err = gWriter.Flush(); err != nil {
		return err
	}
	if err = os.Remove(oldFile); err != nil {
		return err
	}
	return nil
}
