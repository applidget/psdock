package psdock

import (
	"log"
)

func Runner() {
	conf, err := ParseArgs()

	if err != nil {
		log.Fatal(err)
	}

	ps := NewProcess(conf)
	if err = SetUser(ps.Conf.UserName); err != nil {
		log.Fatal(err)
	}
	ps.SetEnvVars()

	ps.Start()

	for {
		status := <-ps.StatusChannel
		if status.Err != nil {
			//Should an error occur, we want to kill the process
			ps.Notif.Notify(PROCESS_STOPPED)
			log.Println(status.Err)
			termErr := ps.Terminate(5)
			log.Println(status.Err)
			log.Println(termErr)
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
