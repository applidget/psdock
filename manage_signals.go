package psdock

import (
	"os"
	"os/signal"
	"syscall"
)

//ManageSignals awaits for incoming signals and triggers a http request when one
//is received. Signals listened to are SIGINT, SIGQUIT, SIGTERM, SIGHUP, SIGALRM and SIGPIPE
func ManageSignals(p *Process) {
	termSignalChannel := make(chan os.Signal)
	otherSignalChannel := make(chan os.Signal)
	signal.Notify(termSignalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGALRM, syscall.SIGPIPE)

	signal.Notify(otherSignalChannel, otherSignals...)
	go func() {
		for {
			s := <-otherSignalChannel
			p.Cmd.Process.Signal(s)
		}
	}()
	_ = <-termSignalChannel

	//Terminate the process and notify
	p.eofChannel <- true
	_ = p.Terminate(5)
}
