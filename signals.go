package psdock

import (
	"os"
	"os/signal"
	"syscall"
)

//ManageSignals awaits for incoming signals and triggers a http request when one
//is received. Signals listened to are SIGINT, SIGQUIT, SIGTERM, SIGHUP, SIGALRM and SIGPIPE
func ManageSignals(p *Process) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGALRM, syscall.SIGPIPE)
	_ = <-signalChannel

	//Terminate the process and notify
	_ = p.Terminate(5)
}
