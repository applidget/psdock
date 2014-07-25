package psdock

import (
	"errors"
	//"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func redirectIO(cmd *exec.Cmd, f *os.File, stdout string, logRotation string, errorChannel chan ProcessStatus) {
	go io.Copy(f, os.Stdin)
	var w io.Writer

	if stdout == "os.Stdout" {
		w = os.Stdout
	} else {
		parsedStdout := strings.Split(stdout, ":")
		prefix := parsedStdout[0]
		if prefix == "" {
			errorChannel <- ProcessStatus{Status: -1, Err: errors.New("Stdout given not supported")}
		}
		if prefix == "file" {
			internalChannel := make(chan error, 1)
			go manageLogRotation(&w, parsedStdout[1][1:], logRotation, internalChannel)
			if err := <-internalChannel; err != nil {
				errorChannel <- ProcessStatus{Status: -1, Err: err}
			}
		} else if prefix == "tcp" {
			//TDB
		}
	}
	io.Copy(w, f)
}

func manageLogRotation(w *io.Writer, path string, logRotation string, errorChannel chan error) {
	var lifetime time.Duration
	switch logRotation {
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
		f, err := os.Create(path + "-" + string(time.Now().Format("Jan-2-2006-3:04pm")))
		if err != nil {
			errorChannel <- err
		}
		*w = f
		errorChannel <- nil
		time.Sleep(lifetime)
	}
}
