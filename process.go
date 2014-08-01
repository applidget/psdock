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
	stdinStruct   *stdin
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
	ppids, _ := getPPIDs(p.Cmd.Process.Pid)
	for _, v := range ppids {
		if v == ownerPid {
			return true
		}
	}
	return false
}

func getPPIDs(pid int) ([]int, error) {
	pgrepCmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	pgrepOutput, _ := pgrepCmd.Output()

	scanner := bufio.NewScanner(strings.NewReader(string(pgrepOutput)))
	ppids := []int{pid}
	for scanner.Scan() {
		childPid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Println(err)
			break
		}
		childPidsRecur, err := getPPIDs(childPid)
		if err != nil {
			log.Println(err)
			break
		}
		ppids = append(ppids, childPidsRecur...)
	}
	return ppids, nil
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
		p.eofChannel = make(chan bool, 1)
		p.stdinStruct, err = setTerminalAndRedirectStdin(os.Stdin, p.Pty)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		defer p.stdinStruct.restoreStdin()

		if err = p.redirectStdout(); err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		refTime := time.Now()
		for !p.isStarted() {
			time.Sleep(100 * time.Millisecond)
			//30s timeout
			if time.Now().Sub(refTime) > 90*time.Second {
				p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED,
					Err: errors.New("Processed did not bind the port within the timeout")}
				return
			}
		}

		if err = p.Notif.Notify(PROCESS_STARTED); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STARTED, Err: nil}

		refTime = time.Now()
		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
			//30s timeout
			if time.Now().Sub(refTime) > 90*time.Second {
				p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED,
					Err: errors.New("Processed did not bind the port within the timeout")}
				return
			}
		}

		if err = p.Notif.Notify(PROCESS_RUNNING); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_RUNNING, Err: nil}

		err = p.Cmd.Wait()
		if err != nil {
			p.Notif.Notify(PROCESS_STOPPED)
			log.Fatal(err)
		}
		_ = <-p.eofChannel

		//p has stopped and stdout has been written
		if err = p.Notif.Notify(PROCESS_STOPPED); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: PROCESS_STOPPED, Err: nil}
	}()

	return nil
}
