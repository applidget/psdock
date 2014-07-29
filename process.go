package psdock

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"github.com/kr/pty"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
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
	Term          *terminal.Terminal
	StatusChannel chan ProcessStatus
	oldTermState  *terminal.State
	output        io.WriteCloser
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
	return &Process{Cmd: cmd, Conf: conf, StatusChannel: newStatusChannel, output: os.Stdout, Notif: Notifier{webHook: conf.WebHook}}
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

//SetUser tries to change the current user to newUsername
func (p *Process) SetUser() error {
	currentUser, err := user.Current()
	if err != nil {
		return errors.New("Can't determine the current user !\n" + err.Error())
	}

	if p.Conf.UserName == currentUser.Username {
		return nil
	}

	newUser, err := user.Lookup(p.Conf.UserName)
	if err != nil {
		return errors.New("Can't find the user" + p.Conf.UserName + "!\n" + err.Error())
	}

	newUserUID, err := strconv.Atoi(newUser.Uid)
	if err != nil {
		return errors.New("Can't determine the new user UID !\n" + err.Error())
	}

	if err := syscall.Setuid(newUserUID); err != nil {
		return errors.New("Can't change the user !\n" + err.Error())
	}

	return nil
}

func (p *Process) Terminate(maxTryCount int) error {
	if maxTryCount > 0 {
		err := syscall.Kill(p.Cmd.Process.Pid, syscall.SIGTERM)
		if err == nil && !p.isRunning() {
			return nil
		}
	} else {
		return syscall.Kill(p.Cmd.Process.Pid, syscall.SIGKILL)
	}
	time.Sleep(100 * time.Millisecond)
	return p.Terminate(maxTryCount - 1)
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
	//We execute netstat -an | grep bindPort to find if bindPort is open
	netstCmd := exec.Command("netstat", "-an")
	grepCmd := exec.Command("grep", string(p.Conf.BindPort))
	netstOut, _ := netstCmd.Output()
	grepCmd.Stdin = bytes.NewBuffer(netstOut)
	grepOut, _ := grepCmd.Output()

	return len(grepOut) > 0
}

func (p *Process) Start() error {
	//We start the process
	var startErr error
	p.Pty, startErr = pty.Start(p.Cmd)
	if startErr != nil {
		return startErr
	}

	go func() {
		var err error
		if err = p.redirectStdin(); err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		if err = p.redirectStdout(); err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
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

		err = p.Cmd.Wait()
		if err != nil {
			p.restoreStdin()
			p.Notif.Notify(PROCESS_STOPPED)
			log.Fatal(err)
		}

		//p has stopped
		p.restoreStdin()
		if err = p.Notif.Notify(PROCESS_STOPPED); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	return nil
}
