package main

import (
	"github.com/applidget/psdock"
	"log"
	"os/exec"
	"strings"
)

func main() {
	statusChannel := make(chan psdock.ProcessStatus, 1)
	Config, err := psdock.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}
	//prepare the process
	var processCmd *exec.Cmd
	if len(Config.Args) > 0 {
		processCmd = exec.Command(Config.Command, strings.Split(Config.Args, " ")...)
	} else {
		processCmd = exec.Command(Config.Command)
	}

	if err := psdock.PrepareProcessEnv(processCmd, Config); err != nil {
		log.Fatal(err)
	}

	//Set up signal monitoring
	go psdock.ManageSignals(processCmd, statusChannel)

	//Launch the process
	go psdock.LaunchProcess(processCmd, Config, statusChannel)

	for {
		code := <-statusChannel
		if code.Err != nil {
			notifyWebHook(Config.WebHook, "stopped")
			log.Fatal(code.Err)
		}
		switch code.Status {
		case psdock.PROCESS_STARTED:
			notifyWebHook(Config.WebHook, "started")
		case psdock.PROCESS_RUNNING:
			notifyWebHook(Config.WebHook, "running")
		case psdock.PROCESS_STOPPED:
			notifyWebHook(Config.WebHook, "stopped")
			return
		}
	}
}

func notifyWebHook(hook, message string) {
	if err := psdock.NotifyWebHook(hook, message); err != nil {
		log.Print(err)
	}
}
