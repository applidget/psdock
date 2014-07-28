package main

import (
	"github.com/applidget/psdock"
	"log"
)

/*aletrnative main using the process type */

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
			termErr := ps.Terminate()
			log.Println(status.Err)
			log.Println(termErr)
			return
		}
		switch status.Status {
		case psdock.PROCESS_STARTED:
			go ps.ManageSignals()
		case psdock.PROCESS_RUNNING:

		case psdock.PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}
