package main

import (
	"fmt"
	"github.com/applidget/psdock"
	"log"
)

func main() {
	conf, err := psdock.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ps := psdock.NewProcess(conf)
	if err = ps.SetUser(); err != nil {
		log.Fatal(err)
	}
	ps.SetEnvVars()

	if err = ps.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		status := <-ps.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			ps.Status = psdock.PROCESS_STOPPED
			ps.NotifyStatusChanged()
			termErr := ps.Terminate(5)
			log.Println(status.Err)
			log.Println(termErr)
			return
		}
		switch status.Status {
		case psdock.PROCESS_STARTED:
			fmt.Println("STAAAAAARTED")
			go psdock.ManageSignals(ps)
		case psdock.PROCESS_RUNNING:

		case psdock.PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}
