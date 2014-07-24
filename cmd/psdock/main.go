package main

import (
	"fmt"
	"github.com/applidget/psdock"
	"log"
	"os/exec"
	"strings"
	"time"
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
	processCmd := exec.Command(arguments.Command, strings.Split(arguments.Args, " ")...)
	if err := psdock.PrepareProcess(processCmd, arguments); err != nil {
		log.Fatal(err)
	}

	//Set up signal monitoring
	//go psdock.ManageSignals(processCmd, arguments.WebHook, statusChannel)

	//Launch the process
	go psdock.LaunchProcess(processCmd, arguments, statusChannel)

	for { //TODO
		code /*, err*/ := <-statusChannel
		/*if err != nil {
			log.Fatal(err)
		}*/
		if code.Err != nil {
			log.Fatal(code.Err)
		}
		switch code.Status {
		case psdock.STARTED:
			notifyOrFail(arguments.WebHook, "started")
		case psdock.RUNNING:
			notifyOrFail(arguments.WebHook, "running")
		case psdock.STOPPED:
			notifyOrFail(arguments.WebHook, "stopped")
			time.Sleep(time.Second * 2) //in order to have time to trigger the hook
			return
		}
	}
}

func notifyOrFail(hook, message string) {
	if err := psdock.NotifyWebHook(hook, message); err != nil {
		log.Print(err)
	}
}

/*func main() {
	c := make(chan int, 1)
	arguments, err := psdock.ParseArguments()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%+v", arguments)
	triggerGiven := len(arguments.WebHook) > 0

	//prepare the process
	processCmd := exec.Command(arguments.Command, strings.Split(arguments.Args, " ")...)
	psdock.SetEnvVars(processCmd, arguments.EnvVars)
	if err = psdock.ChangeUser(arguments.UserName); err != nil {
		log.Fatal(err)
	}

	//Set up monitoring
	if triggerGiven {
		go func() {
			if err1 := psdock.MonitorStart(processCmd, arguments.WebHook, arguments.BindPort); err1 != nil {
				log.Fatal(err1)
			}
			c <- 1
		}()
		//go psdock.ManageSignals(processCmd, arguments.WebHook)
	}

	//Start process
	f, err := pty.Start(processCmd)
	if err != nil {
		log.Fatal("Was not able to start process", err)
	}

	//Will be replaced by a function dealing with logging
	io.Copy(os.Stdout, f)

	_ = <-c
	fmt.Print("ended")
	if err = processCmd.Wait(); err != nil {
		log.Print(err)
	}
	//If we arrive here, that means the process exited by itself.
	//We just signal it to the hook
	if !triggerGiven {
		return
	}

	if err = psdock.NotifyWebHook(arguments.WebHook, "stopped"); err != nil {
		log.Print(err)
	}
}*/
