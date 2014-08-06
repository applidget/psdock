package psdock

import (
	"os"
	"syscall"
)

var otherSignals = []os.Signal{syscall.SIGABRT, syscall.SIGBUS, syscall.SIGCHLD, syscall.SIGCONT, syscall.SIGFPE,
	syscall.SIGILL, syscall.SIGIO, syscall.SIGIOT, syscall.SIGPROF, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGSYS,
	syscall.SIGTRAP, syscall.SIGTSTP, syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGURG, syscall.SIGUSR1,
	syscall.SIGUSR2, syscall.SIGVTALRM, syscall.SIGWINCH, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGEMT}
