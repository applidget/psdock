package psdock

const PROCESS_STARTED int = 0
const PROCESS_RUNNING int = 1
const PROCESS_STOPPED int = 2

//type used to communicate between goroutines
type ProcessStatus struct {
	Status int
	Err    error
}
