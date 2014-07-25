package psdock

import (
	"bytes"
	"errors"
	"github.com/kr/pty"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

//ManageSignals awaits for incoming signals and triggers a http request when one
//is received. Signals listened to are SIGINT, SIGQUIT, SIGTERM, SIGHUP, SIGALRM and SIGPIPE
func ManageSignals(cmd *exec.Cmd, c chan ProcessStatus) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGALRM, syscall.SIGPIPE)
	_ = <-signalChannel

	//We want to kill the process anyway, even if some errors occur when triggering the hook
	termErr := terminateProcess(cmd)

	//Send the request
	c <- ProcessStatus{Status: PROCESS_STOPPED, Err: termErr}
}

//terminateProcess kills the process referenced by cmd.Process
func terminateProcess(cmd *exec.Cmd) error {
	if err := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return errors.New("Failed to kill the process !\n" + err.Error())
	}
	return nil
}

//MonitorStart triggers the hook when the process starts. If bindPort != 0,
//the trigger is also called when bindPort starts to be used.
func LaunchProcess(cmd *exec.Cmd, Config *Config, c chan ProcessStatus) {
	//We start the process
	f, startErr := pty.Start(cmd)
	if startErr != nil {
		c <- ProcessStatus{Status: -1, Err: startErr}
		return
	}

	//TO DELETE
	//Will be replaced by a function dealing with logging
	go redirectIO(cmd, f)
	//go io.Copy(os.Stdout, f)

	startErr = ensureProcessIsStarted(cmd)
	c <- ProcessStatus{Status: PROCESS_STARTED, Err: startErr}

	runErr := ensureProcessIsRunning(cmd, Config.BindPort)
	c <- ProcessStatus{Status: PROCESS_RUNNING, Err: runErr}

	//We wait for the process to end
	runErr = cmd.Wait()
	if runErr != nil {
		c <- ProcessStatus{Status: -1, Err: runErr}
		return
	}
	//If we arrive here, the process ended. We send that info
	c <- ProcessStatus{Status: PROCESS_STOPPED, Err: runErr}
}

//ensureProcessIsStarted returns only after cmd.Process is started
func ensureProcessIsStarted(cmd *exec.Cmd) error {
	//We wait for the process to start
	for cmd.Process == nil {
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

//ensureProcessIsRunning returns only after cmd.Process is up & running
func ensureProcessIsRunning(cmd *exec.Cmd, bindPort int) error {
	//If bindPort is 0 we have nothing to do
	if bindPort == 0 {
		return nil
	}

	//If bindPort is not 0 we wait for bindPort to be used
	for {
		//We execute netstat -an | grep bindPort to find if bindPort is open
		netstCmd := exec.Command("netstat", "-an")
		grepCmd := exec.Command("grep", string(bindPort))
		netstOut, _ := netstCmd.Output()
		grepCmd.Stdin = bytes.NewBuffer(netstOut)
		grepOut, _ := grepCmd.Output()

		if len(grepOut) > 0 {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
}
