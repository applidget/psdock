package psdock

import (
	"log"
	"os"
)

func Runner() {
	log.SetOutput(os.Stdout)
	conf, err := ParseArgs()
	if err != nil {
		log.Fatal("Fatal error in Runner():" + err.Error())
	}
	if conf.Gateway != "" {
		if err := SetGateway(conf.Gateway); err != nil {
			log.Fatal("Fatal error in Runner():" + err.Error())
		}
	}

	ps := NewProcess(conf)
	ps.SetEnvVars()
	ps.Start()

	for {
		status := <-ps.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			ps.Notif.Notify(PROCESS_STOPPED)
			termErr := ps.Terminate(5)
			log.Println("Fatal error in Runner():" + status.Err.Error())
			if termErr != nil {
				log.Println("Error in Runner():Error in Process.Terminate():" + termErr.Error())
			}
			return
		}
		switch status.Status {
		case PROCESS_STARTED:
			go ManageSignals(ps)
		case PROCESS_RUNNING:
		case PROCESS_STOPPED:
			//If we arrive here, process is already stopped, and this has been notified
			return
		}
	}
}
