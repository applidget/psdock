package psdock

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh/terminal"
	"errors"
	"github.com/kr/pty"
	"io"
	"log"
	"net/http"
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
	Pty           *os.File
	Term          *terminal.Terminal
	Status        int
	StatusChannel chan ProcessStatus
	oldTermState  *terminal.State
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
	return &Process{Cmd: cmd, Conf: conf, StatusChannel: newStatusChannel}
}

//SetEnvVars sets the environment variables for the launched process
func (p *Process) SetEnvVars() {
	if len(p.Conf.EnvVars) == 0 {
		//If we do not want to pass any env var to the child process, we still have
		//to write something in p.Cmd.Env so that it is non-empty
		p.Cmd.Env = append(p.Cmd.Env, " ")
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
		if err := syscall.Kill(p.Cmd.Process.Pid, syscall.SIGTERM); err == nil {
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
	return p.isStarted() && p.hasBoundPort()
}

func (p *Process) hasBoundPort() bool {
	if p.Conf.BindPort == 0 {
		return true
	}

	//We execute netstat -an | grep bindPort to find if bindPort is open
	netstCmd := exec.Command("netstat", "-an")
	grepCmd := exec.Command("grep", string(p.Conf.BindPort))
	netstOut, _ := netstCmd.Output()
	grepCmd.Stdin = bytes.NewBuffer(netstOut)
	grepOut, _ := grepCmd.Output()

	return len(grepOut) > 0
}

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
	go io.Copy(os.Stdout, p.Pty)
	return nil
}

func (p *Process) NotifyStatusChanged() error {
	if p.Conf.WebHook == "" {
		return nil
	}
	statusStr := ""
	if p.Status == PROCESS_STARTED {
		statusStr = "started"
	} else if p.Status == PROCESS_RUNNING {
		statusStr = "running"
	} else {
		statusStr = "stopped"
	}
	body := `{
							"ps":
								{ "status":` + statusStr + `}
						}`

	req, err := http.NewRequest("PUT", p.Conf.WebHook, bytes.NewBufferString(body))
	if err != nil {
		return errors.New("Failed to construct the HTTP request" + err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return errors.New("Was not able to trigger the hook!\n" + err.Error())
	}
	return nil
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
		p.Status = PROCESS_STARTED
		if err = p.NotifyStatusChanged(); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: p.Status, Err: nil}

		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
		}
		p.Status = PROCESS_RUNNING
		if err = p.NotifyStatusChanged(); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: p.Status, Err: nil}

		err = p.Cmd.Wait()
		if err != nil {
			p.restoreStdin()
			p.Status = PROCESS_STOPPED
			p.NotifyStatusChanged()
			log.Fatal(err)
		}

		//p has stopped
		p.restoreStdin()
		p.Status = PROCESS_STOPPED
		if err = p.NotifyStatusChanged(); err != nil {
			log.Println(err)
		}
		p.StatusChannel <- ProcessStatus{Status: p.Status, Err: nil}
	}()

	return nil
}
