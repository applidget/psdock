package main

import (
	"github.com/applidget/psdock"
	"github.com/kr/pty"
	"log"
	"os/exec"
	"strings"
)

func main() {
	arguments := psdock.ParseArguments()

	//prepare the process
	processCmd := exec.Cmd(arguments.Process, strings.Split(arguments.Args, " ")...)
	psdock.SetEnvVars(processCmd, arguments.EnvVars)
	if err := psdock.ChangeUser(arguments.UserName); err != nil {
		log.Fatal("Was not able to change the user!", err)
	}
	f, err := pty.Start(processCmd)
	go psdock.MonitorStart(processCmd, arguments.WebHook, arguments.BindPort)
	go psdock.ManageSignals(processCmd, arguments.WebHook)

	if executionError := processCmd.Wait(); executionError != nil {
		log.Print("The process did not work flawlessly : ", executionError)
	}
	//If we arrive here, that means the process exited by itself.
	//We just signal it to the hook
	psdock.SendRequest(processCmd.WebHook, "stopped")
}
