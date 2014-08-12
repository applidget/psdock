package psdock

const EOL rune = 0x0A
const PROCESS_STARTED int = 0
const PROCESS_RUNNING int = 1
const PROCESS_STOPPED int = 2

const PSDOCK_CFG_FILEPATH string = "/etc/psdock/psdock.conf"

//type used to communicate between goroutines
type ProcessStatus struct {
	Status int
	Err    error
}
