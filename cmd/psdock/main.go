package main

import (
	"github.com/applidget/psdock"
	"github.com/kr/pty"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	arguments := psdock.ParseArguments()

	//prepare the process
	processCmd := exec.Command(arguments.Command, strings.Split(arguments.Args, " ")...)
	psdock.SetEnvVars(processCmd, arguments.EnvVars)
	if err := psdock.ChangeUser(arguments.UserName); err != nil {
		log.Fatal("Was not able to change the user!", err)
	}
	f, err := pty.Start(processCmd)
	if err != nil {
		log.Fatal("Was not able to start process")
	}
	go psdock.MonitorStart(processCmd, arguments.WebHook, arguments.BindPort)
	go psdock.ManageSignals(processCmd, arguments.WebHook)

	//Will be replaced by a function dealing with logging
	io.Copy(os.Stdout, f)

	if executionError := processCmd.Wait(); executionError != nil {
		log.Print("The process did not work flawlessly : ", executionError)
	}
	//If we arrive here, that means the process exited by itself.
	//We just signal it to the hook
	psdock.SendRequest(arguments.WebHook, "stopped")
}
