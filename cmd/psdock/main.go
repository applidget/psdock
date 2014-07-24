package main

import (
	"fmt"
	"github.com/applidget/psdock"
	"log"
	"os/exec"
	"strings"
	//"time"
)

func useless() {
	fmt.Printf("TODELETE")
}

func main() {
	statusChannel := make(chan psdock.CommData, 1)
	arguments, err := psdock.ParseArguments()
	if err != nil {
		log.Fatal(err)
	}
	//prepare the process
	var processCmd *exec.Cmd
	if len(arguments.Args) > 0 {
		processCmd = exec.Command(arguments.Command, strings.Split(arguments.Args, " ")...)
	} else {
		processCmd = exec.Command(arguments.Command)
	}

	if err := psdock.PrepareProcess(processCmd, arguments); err != nil {
		log.Fatal(err)
	}

	//Set up signal monitoring
	go psdock.ManageSignals(processCmd, statusChannel)

	//Launch the process
	go psdock.LaunchProcess(processCmd, arguments, statusChannel)

	for {
		code := <-statusChannel
		if code.Err != nil {
			notifyOrFail(arguments.WebHook, "stopped")
			log.Fatal(code.Err)
		}
		switch code.Status {
		case psdock.STARTED:
			notifyOrFail(arguments.WebHook, "started")
		case psdock.RUNNING:
			notifyOrFail(arguments.WebHook, "running")
		case psdock.STOPPED:
			notifyOrFail(arguments.WebHook, "stopped")
			return
		}
	}
}

func notifyOrFail(hook, message string) {
	if err := psdock.NotifyWebHook(hook, message); err != nil {
		log.Print(err)
	}
}
