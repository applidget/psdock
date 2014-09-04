package psdock

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/kr/pty"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Process struct {
	Cmd           *exec.Cmd
	Conf          *Config
	Notif         Notifier
	Pty           *os.File
	ioC           *ioContext
	StatusChannel chan ProcessStatus
	eofChannel    chan bool
}

//NewProcess creates a new struct of type *Process and returns its address
func NewProcess(conf *Config) *Process {
	var cmd *exec.Cmd
	shell := os.Getenv("SHELL")
	cmd = exec.Command(shell, "-c", conf.Command)
	newStatusChannel := make(chan ProcessStatus, 1)

	return &Process{Cmd: cmd, Conf: conf, StatusChannel: newStatusChannel, Notif: Notifier{webHook: conf.WebHook}}
}

//SetEnvVars sets the environment variables for the launched process
//If p.Conf.EnvVars is empty, we pass all the current env vars to the child
func (p *Process) SetEnvVars() {
	if len(p.Conf.EnvVars) == 0 {
		return
	}
	for _, envVar := range strings.Split(p.Conf.EnvVars, ",") {
		p.Cmd.Env = append(p.Cmd.Env, envVar)
	}
}

//Terminate sends a SIGTERM signal, then waits nbSec seconds before sending a SIGKILL if necessary.
//p.ioC.restoreIO() will be called at the end of the function
func (p *Process) Terminate(nbSec int) error {
	defer p.ioC.restoreIO()
	if !p.isRunning() {
		return nil
	}
	syscall.Kill(p.Cmd.Process.Pid, syscall.SIGTERM)

	time.Sleep(time.Second)
	if !p.isRunning() {
		return nil
	}
	time.Sleep(time.Duration(nbSec) * time.Second)
	if !p.isRunning() {
		return nil
	}
	return syscall.Kill(p.Cmd.Process.Pid, syscall.SIGKILL)
}

//returns true if p is started
func (p *Process) isStarted() bool {
	return p.Cmd.Process != nil
}

//returns true if p is running
func (p *Process) isRunning() bool {
	if p.Conf.BindPort == 0 {
		return p.isStarted()
	} else {
		return p.isStarted() && p.hasBoundPort()
	}
}

//hasBoundPort returns true if p has bound its port
func (p *Process) hasBoundPort() bool {
	//We execute lsof -i :bindPort to find if bindPort is open
	lsofCmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(p.Conf.BindPort))
	lsofBytes, _ := lsofCmd.Output()
	return parseLsof(lsofBytes, p.Cmd.Process.Pid, getPIDs)
}

func parseLsof(lsofBytes []byte, pid int, retrievePIDs func(int) ([]int, error)) bool {
	lsofScanner := bufio.NewScanner(bytes.NewBuffer(lsofBytes))
	lsofScanner.Scan()
	lsofScanner.Text()
	lsofScanner.Scan()
	lsofResult := lsofScanner.Text()
	if len(lsofResult) == 0 {
		return false
	}

	plsofResult := strings.Split(lsofResult, "    ")

	plsofResult = strings.Split(plsofResult[1], " ")
	ownerPid, _ := strconv.Atoi(plsofResult[0])
	ppids, _ := retrievePIDs(pid)
	for _, v := range ppids {
		if v == ownerPid {
			return true
		}
	}
	return false
}

func (p *Process) Start() {
	initCompleteChannel := make(chan bool)
	runningMessChannel := make(chan bool)
	p.eofChannel = make(chan bool, 1)

	go func() {
		var startErr error

		//For the moment, the go implementation of the setuid syscall is buggy : it only affects the current thread, not the whole process
		//(see https://code.google.com/p/go/issues/detail?id=1435)
		//Therefore, the call has to be done in the goroutine used to launch the process
		//The setuser call should be moved in runner.go (before the Start() call) once this bug is solved
		if err := SetUser(p.Conf.UserName); err != nil {
			log.Fatal("Fatal error in Runner():" + err.Error())
		}

		p.Pty, startErr = pty.Start(p.Cmd)
		if startErr != nil {
			log.Printf("%#v", p.Cmd)
			p.ioC.restoreIO()
			p.StatusChannel <- ProcessStatus{Status: -1, Err: errors.New("Error in Process.Start():Error in pty.Start():" + startErr.Error())}
		}
		initCompleteChannel <- true
		_ = <-runningMessChannel
		err := p.Cmd.Wait()
		if err != nil {
			log.Println("Error in Process.start():Error in p.Cmd.Wait():" + err.Error())
			p.ioC.restoreIO()
			p.Notif.Notify(PROCESS_STOPPED)
			p.Terminate(5)
		}
		_ = <-p.eofChannel
		if err = p.Notif.Notify(PROCESS_STOPPED); err != nil {
			log.Println("Error in Process.start():" + err.Error())
		}
		p.ioC.restoreIO()
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	go func() {
		var err error
		_ = <-initCompleteChannel
		p.ioC, err = newIOContext(p.Conf.Stdin, p.Pty, p.Conf.Stdout, p.Conf.LogPrefix, p.Conf.LogRotation, p.Conf.LogColor,
			p.StatusChannel, p.eofChannel)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: errors.New("Error in Process.start():" + err.Error())}
		}

		for !p.isStarted() {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_STARTED); err != nil {
			log.Println("Error in Process.start():" + err.Error())
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STARTED, Err: nil}
		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
		}
		if err = p.Notif.Notify(PROCESS_RUNNING); err != nil {
			log.Println("Error in Process.start():" + err.Error())
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_RUNNING, Err: nil}
		runningMessChannel <- true
	}()
}
