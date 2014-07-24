package psdock

const STARTED int = 0
const RUNNING int = 1
const STOPPED int = 2

//type used to communicate between goroutines
type CommData struct {
	Status int
	Err    error
}
