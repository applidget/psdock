package psdock

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

//ManageSignals awaits for incoming signals and triggers a http request when one
//is received. Signals listened to are SIGINT, SIGQUIT, SIGTERM, SIGHUP, SIGALRM and SIGPIPE
func ManageSignals(cmd *exec.Cmd, hook string) error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGALRM, syscall.SIGPIPE)
	_ = <-c

	//We want to kill the process anyway, even if some errors occur when triggering the hook
	defer terminateProcess(cmd)

	//Send the request
	SendRequest(hook, "stopped")
}

//killProcess kills the process referenced by cmd.Process
func terminateProcess(cmd *exec.Cmd) error {
	if err := syscall.Kill(cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return errors.New("Failed to kill the process !\n"+err)
	}
}

//MonitorStart triggers the hook when the process starts. If bindPort != 0,
//the trigger is also called when bindPort starts to be used.
func MonitorStart(cmd *exec.Cmd, hook string, bindPort int) {
	const harmelessSignalIndex := 0
	//We wait for the process to exist
	for cmd.Process != nil {
		time.Sleep(100 * time.Millisecond)
	}

	//We send the signal 0 to the process. It doesn't do anything, but we can still
	//check the error returned by Process.Signal. If it is nil, the process is running
	err := cmd.Process.Signal(syscall.Signal(0))
	for err != nil {
		log.Print("The process doesn't seem to be running")
		time.Sleep(3 * time.Second)
		err = cmd.Process.Signal(syscall.Signal(harmelessSignalIndex))
	}

	//if bindPort is 0, we send a "running" message
	if bindPort == 0 {
		SendRequest(hook, "running")
		return
	}
	//if bindPort is not 0, we send a "started" message
	SendRequest(hook, "started")

	//We wait for bindPort to be used, and then send a "running" message
	for {
		//We execute netstat -an | grep bindPort to find if bindPort is open
		netstCmd := exec.Command("netstat", "-an")
		grepCmd := exec.Command("grep", string(bindPort))
		netstOut, _ := netstCmd.Output()
		grepCmd.Stdin = bytes.NewBuffer(netstOut)
		grepOut, _ := grepCmd.Output()

		if len(grepOut) > 0 {
			SendRequest(hook, "running")
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}

//sendRequest sends a http "PUT" request to hook. The message is of type json, and
//is "{"ps":{"status":status}}
func NotifyWebHook(hook string, status string) error {
	requestMessage := strings.Join([]string{"{\"ps\":{\"status\":", status, "}}"}, "")
	request, err := http.NewRequest("PUT", hook, bytes.NewBufferString(requestMessage))
	if err != nil {
		return errors.New("Failed to contruct the HTTP request.\n"+err)
	}
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	//Send the request
	if _, err := client.Do(request); err != nil {
		return errors.New("Was not able to send the request : "+request+"\n"+err)
	}
}
