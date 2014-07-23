package psdock

import (
	"flag"
	"log"
)

//command is the name of the command to be executed by psdock
//arguments contains the result of command-line-arguments parsing.
//args are the arguments passed to this command
//stdout is the redirection path for the stdout/stderr of the launched process
//logRotation is the lifetime (in seconds) of a single log file
//logPrefix is the prefix for logging the output of the launched process
//watchPort is the port watched for binding by psdock
//httpHook is the hook triggered by psdock in case of special events
//user is the UID of the user launching the process
type arguments struct {
	command     string
	args        string
	stdout      string
	logRotation string
	logPrefix   string
	envVars     string
	watchPort   int
	httpHook    string
	userUID     int
}

//ParseArguments parses command-line arguments and returns them in an arguments struct
func ParseArguments() arguments {
	parsedArgs := arguments{}
	flag.StringVar(&parsedArgs.command, "command", "", "name of the command to be executed by psdock")
	flag.StringVar(&parsedArgs.args, "args", "", "arguments passed to the launched command")
	flag.StringVar(&parsedArgs.stdout, "stdout", "os.Stdout", "redirection path for the stdout/stderr of the launched process")
	flag.StringVar(&parsedArgs.logRotation, "logRotation", "daily", "lifetime of a single log file.")
	flag.StringVar(&parsedArgs.logPrefix, "logPrefix", "", "prefix for logging the output of the launched process")
	flag.StringVar(&parsedArgs.envVars, "envVars", "", "arguments passed to the launched command")
	flag.IntVar(&parsedArgs.watchPort, "watchPort", 0, "port to be watched for binding by psdock")
	flag.StringVar(&parsedArgs.httpHook, "httpHook", "", "hook triggered by psdock in case of special events")
	flag.IntVar(&parsedArgs.userUID, "user", 0, "UID of the user launching the process")

	flag.Parse()

	if parsedArgs.command == "" {
		flag.PrintDefaults()
		log.Fatal("You must give a command to start !")
	}

	if parsedArgs.stdout == "" {
		flag.PrintDefaults()
		log.Fatal("stdout can't be nil !")
	}

	if parsedArgs.logRotation != "minutely" && parsedArgs.logRotation != "hourly" &&
		parsedArgs.logRotation != "daily" && parsedArgs.logRotation != "weekly" {
		flag.PrintDefaults()
		log.Fatal("logRotation has to be minutely, hourly, daily or weekly !")
	}
	if parsedArgs.watchPort > 0 && parsedArgs.httpHook == "" {
		flag.PrintDefaults()
		log.Fatal("If you specify a port, you have to specify a http hook !")
	}
	if parsedArgs.watchPort < 0 {
		flag.PrintDefaults()
		log.Fatal("watchPort can't be negative!")
	}
	if parsedArgs.userUID < 0 {
		flag.PrintDefaults()
		log.Fatal("userUID can't be negative!")
	}

	return parsedArgs
}
