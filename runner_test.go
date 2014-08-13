package psdock

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
	"time"
)

func runnerForTesting(conf *Config) {
	var err error
	SetGateway(conf.Gateway)
	p := NewProcess(conf)
	if err = SetUser(p.Conf.UserName); err != nil {
		log.Fatal(err)
	}
	p.SetEnvVars()

	initCompleteChannel := make(chan bool)
	runningMessChannel := make(chan bool)
	p.eofChannel = make(chan bool, 1)

	go func() {
		var startErr error
		p.Pty, startErr = pty.Start(p.Cmd)
		if startErr != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: startErr}
		}
		initCompleteChannel <- true
		_ = <-runningMessChannel
		err := p.Cmd.Wait()

		if err != nil {
			log.Println(err)
			p.Notif.Notify(PROCESS_STOPPED)
			p.Terminate(5)
		}
		_ = <-p.eofChannel
		if err = p.Notif.Notify(PROCESS_STOPPED); err != nil {
			log.Println(err)
		}
		p.ioC.restoreIO()
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	go func() {
		var err error
		_ = <-initCompleteChannel

		p.ioC = &ioContext{}
		err = p.ioC.redirectStdout(p.Pty, p.Conf.Stdout, p.Conf.LogPrefix, p.Conf.LogRotation, p.Conf.LogColor, p.StatusChannel, p.eofChannel)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		pty := p.Pty
		stdinStr := p.Conf.Stdin
		statusChannel := p.StatusChannel

		url, err := url.Parse(stdinStr)
		if err != nil {
			log.Println(err)
		}
		if url.Path == "os.Stdin" {
			//We don't need to do anything here
		} else if url.Scheme == "tcp" {
			conn, err := net.Dial("tcp", url.Host+url.Path)
			if err != nil {
				log.Println(err)
			}
			//Directly copy from the connection to the pty. Escape chars won't be available
			go func() {
				io.Copy(pty, conn)
				//When the remote stdin closes, terminate the process through the status Channel
				statusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: errors.New("Remote stdin closed")}
			}()
		} else {
			//default case, the protocol is not supported
			log.Println("The protocol " + url.Scheme + " is not supported")
		}
		p.ioC.stdinOutput = os.Stdin

		if err != nil {
			log.Println("Can't create terminal:" + err.Error())
		}
		p.ioC.term = terminal.NewTerminal(p.ioC.stdinOutput, "")
		cb := func(s string, i int, r rune) (string, int, bool) {
			car := []byte{byte(r)}
			p.ioC.term.Write(car)
			return s, i, false
		}
		p.ioC.term.AutoCompleteCallback = cb

		//Copy everything from os.Stdin to the pty
		go io.Copy(pty, p.ioC.stdinOutput)

		//Write all the color symbols
		colors := []string{"magenta", "white", "red", "blue", "green", "yellow", "cyan", "black"}
		for _, color := range colors {
			p.ioC.setTerminalColor(color)
		}

		for !p.isStarted() {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_STARTED); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STARTED, Err: nil}

		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_RUNNING); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_RUNNING, Err: nil}
		runningMessChannel <- true
	}()

	for {
		status := <-p.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			p.Notif.Notify(PROCESS_STOPPED)
			log.Println(status.Err)
			termErr := p.Terminate(5)
			log.Println(status.Err)
			log.Println(termErr)
			return
		}
		switch status.Status {
		case PROCESS_STARTED:
			go ManageSignals(p)
		case PROCESS_RUNNING:

		case PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}

func TestPsdock(t *testing.T) {
	//Before running this test, /etc/psdock/psdock.conf should be :
	//Command = "ls"
	cmd := exec.Command("ls")
	boutLs, err := cmd.Output()
	outLs := string(boutLs)
	outLs = strings.Replace(outLs, "\n", " ", -1)
	if err != nil {
		log.Fatal(err)
	}
	//Redirect the stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"psdock", "--command", "ls"}
	conf, err := ParseArgs()

	if err != nil {
		log.Fatal(err)
	}
	runnerForTesting(conf)

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oldStdout
	outPsdock := <-outC

	lsScanner := bufio.NewScanner(strings.NewReader(outLs))
	lsScanner.Split(bufio.ScanWords)
	lsSpliceResult := []string{}
	for lsScanner.Scan() {
		lsSpliceResult = append(lsSpliceResult, lsScanner.Text())
	}
	sort.Sort(sort.StringSlice(lsSpliceResult))
	lsSpliceResultStr := strings.Join(lsSpliceResult, "")
	psdockScanner := bufio.NewScanner(strings.NewReader(outPsdock))
	psdockScanner.Split(bufio.ScanWords)
	psdockSpliceResult := []string{}
	for psdockScanner.Scan() {
		psdockSpliceResult = append(psdockSpliceResult, psdockScanner.Text())
	}
	sort.Sort(sort.StringSlice(psdockSpliceResult))
	psdockSpliceResultStr := strings.Join(psdockSpliceResult, "")
	if lsSpliceResultStr != psdockSpliceResultStr {
		t.Error("expected:" + lsSpliceResultStr + "\n\ngot:" + psdockSpliceResultStr)
	}
}

func TestPsdockFileOutput(t *testing.T) {
	s := "file://mylog"
	conf := &Config{Command: "ls", Args: "", Stdout: s, LogRotation: "daily", LogColor: "black", LogPrefix: "[PRFX]",
		EnvVars: "MYKEY = myval", BindPort: 0, Stdin: "os.Stdin", UserName: "", WebHook: "http://www.google.fr", Gateway: "10.0.3.1"}

	runnerForTesting(conf)
	fn, _ := retrieveFilenames("mylog", ".log")
	s = fn[0]
	psdockBytes, err := ioutil.ReadFile(s[2:])
	if err != nil {
		t.Error(err)
	}
	cmd := exec.Command("ls")
	os.Remove(s[2:])
	boutLs, _ := cmd.Output()
	outLs := string(boutLs)
	outLs = strings.Replace(outLs, "\n", " ", -1)

	psdockStr := string(psdockBytes)
	psdockStr = strings.Replace(psdockStr, s[2:], "", -1)

	lsScanner := bufio.NewScanner(strings.NewReader(outLs))
	lsScanner.Split(bufio.ScanWords)
	lsSpliceResult := []string{}
	for lsScanner.Scan() {
		lsSpliceResult = append(lsSpliceResult, lsScanner.Text())
	}
	sort.Sort(sort.StringSlice(lsSpliceResult))
	lsSpliceResultStr := strings.Join(lsSpliceResult, "")

	psdockScanner := bufio.NewScanner(strings.NewReader(psdockStr))
	psdockScanner.Split(bufio.ScanWords)
	psdockSpliceResult := []string{}
	for psdockScanner.Scan() {
		psdockSpliceResult = append(psdockSpliceResult, psdockScanner.Text())
	}
	sort.Sort(sort.StringSlice(psdockSpliceResult))
	psdockSpliceResultStr := strings.Join(psdockSpliceResult, "")
	lsSpliceResultStr = strings.Replace(lsSpliceResultStr, s[2:], "", -1)
	if psdockSpliceResultStr != lsSpliceResultStr {
		t.Error("Expected" + lsSpliceResultStr + ", got:" + psdockSpliceResultStr)
	}
}

func runnerForTestingWithTCPOutput(conf *Config) {
	var err error

	p := NewProcess(conf)
	if err = SetUser(p.Conf.UserName); err != nil {
		log.Fatal(err)
	}
	p.SetEnvVars()

	initCompleteChannel := make(chan bool)
	runningMessChannel := make(chan bool)
	p.eofChannel = make(chan bool, 1)

	go func() {
		var startErr error
		p.Pty, startErr = pty.Start(p.Cmd)
		if startErr != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: startErr}
		}
		initCompleteChannel <- true
		_ = <-runningMessChannel
		err := p.Cmd.Wait()
		p.ioC.log.output.Close()
		if err != nil {
			log.Println(err)
			p.Notif.Notify(PROCESS_STOPPED)
			p.Terminate(5)
		}
		_ = <-p.eofChannel
		if err = p.Notif.Notify(PROCESS_STOPPED); err != nil {
			log.Println(err)
		}
		p.ioC.restoreIO()
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	go func() {
		var err error
		_ = <-initCompleteChannel

		p.ioC = &ioContext{}
		err = p.ioC.redirectStdout(p.Pty, p.Conf.Stdout, p.Conf.LogPrefix, p.Conf.LogRotation, p.Conf.LogColor, p.StatusChannel, p.eofChannel)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		pty := p.Pty
		stdinStr := p.Conf.Stdin
		statusChannel := p.StatusChannel

		url, err := url.Parse(stdinStr)
		if err != nil {
			log.Println(err)
		}
		if url.Path == "os.Stdin" {
			//We don't need to do anything here
		} else if url.Scheme == "tcp" {
			conn, err := net.Dial("tcp", url.Host+url.Path)
			if err != nil {
				log.Println(err)
			}
			//Directly copy from the connection to the pty. Escape chars won't be available
			go func() {
				io.Copy(pty, conn)
				//When the remote stdin closes, terminate the process through the status Channel
				statusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: errors.New("Remote stdin closed")}
			}()
		} else {
			//default case, the protocol is not supported
			log.Println("The protocol " + url.Scheme + " is not supported")
		}
		p.ioC.stdinOutput = os.Stdin

		if err != nil {
			log.Println("Can't create terminal:" + err.Error())
		}
		p.ioC.term = terminal.NewTerminal(p.ioC.stdinOutput, "")
		cb := func(s string, i int, r rune) (string, int, bool) {
			car := []byte{byte(r)}
			p.ioC.term.Write(car)
			return s, i, false
		}
		p.ioC.term.AutoCompleteCallback = cb

		//Copy everything from os.Stdin to the pty
		go io.Copy(pty, p.ioC.stdinOutput)

		for !p.isStarted() {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_STARTED); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STARTED, Err: nil}

		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_RUNNING); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_RUNNING, Err: nil}
		runningMessChannel <- true
	}()

	for {
		status := <-p.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			p.Notif.Notify(PROCESS_STOPPED)
			log.Println(status.Err)
			termErr := p.Terminate(5)
			log.Println(status.Err)
			log.Println(termErr)
			return
		}
		switch status.Status {
		case PROCESS_STARTED:
			go ManageSignals(p)
		case PROCESS_RUNNING:

		case PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}

func TestPsdockTCPOutput(t *testing.T) {
	cmd := exec.Command("ls")
	boutLs, _ := cmd.Output()
	outLs := string(boutLs)
	outLs = strings.Replace(outLs, "\n", " ", -1)
	s := "tcp://127.0.0.1:8080"
	nc := exec.Command("nc", "-l", "8080")
	stdout, _ := nc.StdoutPipe()
	nc.Start()
	time.Sleep(time.Second)
	conf := Config{Command: "ls", Args: "", Stdout: s, LogRotation: "daily", LogColor: "black", LogPrefix: "",
		EnvVars: "", BindPort: 0, Stdin: "os.Stdin", UserName: "", WebHook: "http://www.google.fr", Gateway: "10.0.3.1"}
	runnerForTestingWithTCPOutput(&conf)

	psdockStr, _ := ioutil.ReadAll(stdout)
	lsScanner := bufio.NewScanner(strings.NewReader(outLs))
	lsScanner.Split(bufio.ScanWords)
	lsSpliceResult := []string{}
	for lsScanner.Scan() {
		lsSpliceResult = append(lsSpliceResult, lsScanner.Text())
	}
	sort.Sort(sort.StringSlice(lsSpliceResult))
	lsSpliceResultStr := strings.Join(lsSpliceResult, "")

	psdockScanner := bufio.NewScanner(strings.NewReader(string(psdockStr)))
	psdockScanner.Split(bufio.ScanWords)
	psdockSpliceResult := []string{}
	for psdockScanner.Scan() {
		psdockSpliceResult = append(psdockSpliceResult, psdockScanner.Text())
	}
	sort.Sort(sort.StringSlice(psdockSpliceResult))
	psdockSpliceResultStr := strings.Join(psdockSpliceResult, "")
	lsSpliceResultStr = strings.Replace(lsSpliceResultStr, s[2:], "", -1)
	if psdockSpliceResultStr != lsSpliceResultStr {
		t.Error("Expected" + lsSpliceResultStr + ", got:" + psdockSpliceResultStr)
	}
}

func TestLogRotate(t *testing.T) {
	s := "file://mylog"
	conf := &Config{Command: "sleep", Args: "70", Stdout: s, LogRotation: "minutely", LogColor: "black", LogPrefix: "",
		EnvVars: "", BindPort: 0, Stdin: "os.Stdin", UserName: "", WebHook: "", Gateway: "10.0.3.1"}

	runnerForTesting(conf)
	fn, _ := retrieveFilenames("mylog", ".gz")
	s = fn[0]
	os.Remove(s[2:])
}
