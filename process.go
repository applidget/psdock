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
	ioInfo        *ioStruct
	StatusChannel chan ProcessStatus
	eofChannel    chan bool
}

//NewProcess creates a new struct of type *Process and returns its address
func NewProcess(conf *Config) *Process {
	var cmd *exec.Cmd
	if len(conf.Args) > 0 {
		cmd = exec.Command(conf.Command, strings.Split(conf.Args, " ")...)
	} else {
		cmd = exec.Command(conf.Command)
	}
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

func (p *Process) Terminate(nbSec int) error {
	syscall.Kill(p.Cmd.Process.Pid, syscall.SIGTERM)
	time.Sleep(time.Duration(nbSec) * time.Second)
	if !p.isRunning() {
		return nil
	}
	return syscall.Kill(p.Cmd.Process.Pid, syscall.SIGKILL)
}

func (p *Process) isStarted() bool {
	return p.Cmd.Process != nil
}

func (p *Process) isRunning() bool {
	if p.Conf.BindPort == 0 {
		return p.isStarted()
	} else {
		return p.isStarted() && p.hasBoundPort()
	}
}

func (p *Process) hasBoundPort() bool {
	//We execute lsof -i :bindPort to find if bindPort is open
	//For the moment, we only verified that bindPort is used by some process
	lsofCmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(p.Conf.BindPort))

	lsofBytes, _ := lsofCmd.Output()
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
	ppids, _ := getPIDs(p.Cmd.Process.Pid)
	for _, v := range ppids {
		if v == ownerPid {
			return true
		}
	}
	return false
}

func (p *Process) Start() {
	initCompleteChannel := make(chan bool)
	p.eofChannel = make(chan bool, 1)

	go func() {
		var startErr error
		p.Pty, startErr = pty.Start(p.Cmd)
		if startErr != nil {
			log.Println(startErr)
		}
		initCompleteChannel <- true

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
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	go func() {
		var err error
		_ = <-initCompleteChannel

		p.ioInfo, err = newIOStruct(os.Stdin, p.Pty, p.Conf.Stdout, p.Conf.LogPrefix, p.Conf.LogRotation, p.Conf.LogColor,
			p.StatusChannel, p.eofChannel)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		defer p.ioInfo.restoreIO()

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
	}()
}
