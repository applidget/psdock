package psdock

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

//ManageSignals awaits for incoming signals and triggers a http request when one
//is received. Signals listened to are SIGINT, SIGQUIT, SIGTERM, SIGHUP, SIGALRM and SIGPIPE
func ManageSignals(cmd *exec.Cmd, c chan CommData) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGALRM, syscall.SIGPIPE)
	_ = <-signalChannel

	//We want to kill the process anyway, even if some errors occur when triggering the hook
	termErr := terminateProcess(cmd)

	//Send the request
	c <- CommData{Status: STOPPED, Err: termErr}
}

//killProcess kills the process referenced by cmd.Process
func terminateProcess(cmd *exec.Cmd) error {
	if err := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return errors.New("Failed to kill the process !\n" + err.Error())
	}
	return nil
}

//MonitorStart triggers the hook when the process starts. If bindPort != 0,
//the trigger is also called when bindPort starts to be used.
func LaunchProcess(cmd *exec.Cmd, arguments *Arguments, c chan CommData) {
	//We start the process
	f, startErr := pty.Start(cmd)
	if startErr != nil {
		c <- CommData{Status: -1, Err: startErr}
		return
	}

	//TO DELETE
	//Will be replaced by a function dealing with logging
	io.Copy(os.Stdout, f)

	startErr = ensureProcessIsStarted(cmd)
	c <- CommData{Status: STARTED, Err: startErr}

	runErr := ensureProcessIsRunning(cmd, arguments.BindPort)
	c <- CommData{Status: RUNNING, Err: runErr}

	//We wait for the process to end
	runErr = cmd.Wait()
	if runErr != nil {
		c <- CommData{Status: -1, Err: runErr}
		return
	}
	//If we arrive here, the process ended. We send that info
	c <- CommData{Status: STOPPED, Err: runErr}
}

//ensureProcessIsStarted returns only after cmd.Process is started
func ensureProcessIsStarted(cmd *exec.Cmd) error {
	const harmelessSignalIndex int = 0
	//We wait for the process to exist
	for cmd.Process == nil {
		time.Sleep(100 * time.Millisecond)
	}

	//We send the signal 0 to the process. It doesn't do anything, but we can still
	//check the error returned by Process.Signal. If it is nil, the process is running
	err := cmd.Process.Signal(syscall.Signal(harmelessSignalIndex))
	for err != nil {
		fmt.Print("The process doesn't seem to be running")
		time.Sleep(3 * time.Second)
		err = cmd.Process.Signal(syscall.Signal(harmelessSignalIndex))
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
