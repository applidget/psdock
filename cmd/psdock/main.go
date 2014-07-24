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
	arguments, err := psdock.ParseArguments()
	if err != nil {
		log.Fatal(err)
	}

	//prepare the process
	processCmd := exec.Command(arguments.Command, strings.Split(arguments.Args, " ")...)
	psdock.SetEnvVars(processCmd, arguments.EnvVars)
	if err = psdock.ChangeUser(arguments.UserName); err != nil {
		log.Fatal(err)
	}
	f, err := pty.Start(processCmd)
	if err != nil {
		log.Fatal("Was not able to start process")
	}
	go psdock.MonitorStart(processCmd, arguments.WebHook, arguments.BindPort)
	go psdock.ManageSignals(processCmd, arguments.WebHook)

	//Will be replaced by a function dealing with logging
	io.Copy(os.Stdout, f)

	if err = processCmd.Wait(); err != nil {
		log.Print(err)
	}
	//If we arrive here, that means the process exited by itself.
	//We just signal it to the hook
	if err = psdock.NotifyWebHook(arguments.WebHook, "stopped"); err != nil {
		log.Print(err)
	}
}
