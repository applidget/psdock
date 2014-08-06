package main

import (
	"github.com/applidget/psdock"
	"log"
)

func main() {
	if err := psdock.SetNetwork(); err != nil {
		log.Fatal(err)
	}
	conf, err := psdock.ParseArgs()

	if err != nil {
		log.Fatal(err)
	}

	ps := psdock.NewProcess(conf)
	if err = psdock.SetUser(ps.Conf.UserName); err != nil {
		log.Fatal(err)
	}
	ps.SetEnvVars()

	ps.Start()

	for {
		status := <-ps.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			ps.Notif.Notify(psdock.PROCESS_STOPPED)
			log.Println(status.Err)
			termErr := ps.Terminate(5)
			log.Println(status.Err)
			log.Println(termErr)
			return
		}
		switch status.Status {
		case psdock.PROCESS_STARTED:
			go psdock.ManageSignals(ps)
		case psdock.PROCESS_RUNNING:

		case psdock.PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}