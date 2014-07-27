package psdock

import (
	"code.google.com/p/go.crypto/ssh/terminal"
	"exec"
	"os"
	"os/user"
)

type Process struct {
	Cmd    *exec.Cmd
	Conf   *Config
	Pty    *os.File
	Term   *terminal.Terminal
	Status string
}

func NewProcess(conf *Config) *Process {
	var cmd *exec.Cmd
	if len(conf.Args) > 0 {
		cmd = exec.Command(conf.Command, strings.Split(con.Args, " ")...)
	} else {
		cmd = exec.Command(conf.Command)
	}
	return &Process{Cmd: cmd, Conf: conf}
}

func (p *Process) SetEnvVars() {
	if len(p.Conf.EnvVars) == 0 {
		return
	}
	for _, envVar := range strings.Split(p.Conf.EnvVars, ",") {
		p.Cmd.Env = append(p.Cmd.Env, envVar)
	}
}

func (p *Process) SetUser() err {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	if p.Conf.UserName == currentUser.Username {
		return nil
	}

	newUser, err := user.Lookup(newUsername)
	if err != nil {
		return err
	}

	newUserUID, err := strconv.Atoi(newUser.Uid)
	if err != nil {
		return err
	}

	if err := syscall.Setuid(newUserUID); err != nil {
		return err
	}

	return nil
}

func (p *Process) Terminate() err {
	if err := syscall.Kill(p.Cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return err
	}
	return nil
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

func (p *Process) redirectStdin() {

}

func (p *Process) restoreStdin() {

}

func (p *Process) redirectStdout() {

}

func (p *Process) notifyStatusChanged() err {
	if p.Conf.WebHook == "" {
		return nil
	}

	body := `{
							"ps":
								{ "status":`, p.Status, `}
						}`

	req, err := http.NewRequest("PUT", p.Conf.WebHook, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Process) Start(c chan string) error {
	//We start the process
	var err error
	p.Pty, err = pty.Start(cmd)
	if err != nil {
		return err
	}

	go func() {
		p.RedirectStdin()
		p.RedirectStdou()

		for p.isStarted() == false {
			time.Sleep(100 * time.Millisecond)
		}
		p.Status = PROCESS_STARTED
		p.notifyStatusChanged()
		c <- p.Status

		for p.isRunning() == false {
			time.Sleep(100 * time.Millisecond)
		}
		p.Status = PROCESS_RUNNING
		p.notifyStatusChanged()
		c <- p.Status

		cmd.Wait()

		//p has stopped
		p.restoreStdin()
		p.Status = PROCESS_STOPPED
		p.notifyStatusChanged()
		c <- p.Status
	}()
}
