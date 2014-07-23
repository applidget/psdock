package psdock

import (
	"flag"
	"log"
	"os/user"
	"strings"
)

//command is the name of the command to be executed by psdock
//arguments contains the result of command-line-arguments parsing.
//args are the arguments passed to this command
//stdout is the redirection path for the stdout/stderr of the launched process
//logRotation is the lifetime (in seconds) of a single log file
//logPrefix is the prefix for logging the output of the launched process
//bindPort is the port watched for binding by psdock
//webHook is the hook triggered by psdock in case of special events
//user is the UID of the user launching the process
type arguments struct {
	command     string
	args        string
	stdout      string
	logRotation string
	logPrefix   string
	envVars     string
	bindPort    int
	webHook     string
	userName    string
}

//ParseArguments parses command-line arguments and returns them in an arguments struct
func ParseArguments() arguments {
	parsedArgs := arguments{}

	flag.StringVar(&parsedArgs.command, "-process", "", "process to be executed by psdock")
	flag.StringVar(&parsedArgs.stdout, "-stdout", "os.Stdout", "redirection path for the stdout/stderr of the launched process")
	flag.StringVar(&parsedArgs.logRotation, "-log-rotation", "daily", "lifetime of a single log file.")
	flag.StringVar(&parsedArgs.logPrefix, "-log-prefix", "", "prefix for logging the output of the launched process")
	flag.StringVar(&parsedArgs.envVars, "-envVars", "", "arguments passed to the launched command")
	flag.IntVar(&parsedArgs.bindPort, "-bind-port", 0, "port to be watched for binding by psdock")
	flag.StringVar(&parsedArgs.webHook, "-web-hook", "", "hook triggered by psdock in case of special events")

	//The user has to specify a process to run
	if parsedArgs.command == "" {
		flag.PrintDefaults()
		log.Fatal("You must provide a process to run!")
	}
	commandSplited := strings.SplitAfterN(parsedArgs.command, " ", 2)
	parsedArgs.command = commandSplited[0]
	if len(commandSplited) == 1 {
		parsedArgs.args = ""
	} else {
		parsedArgs.args = commandSplited[1]
	}
	//Retrieve the name of the current user
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&parsedArgs.userName, "user-name", user.Username, "name of the user launching the process")

	flag.Parse()

	if parsedArgs.stdout == "" {
		flag.PrintDefaults()
		log.Fatal("stdout can't be nil !")
	}
	if parsedArgs.logRotation != "minutely" && parsedArgs.logRotation != "hourly" &&
		parsedArgs.logRotation != "daily" && parsedArgs.logRotation != "weekly" {
		flag.PrintDefaults()
		log.Fatal("logRotation has to be minutely, hourly, daily or weekly !")
	}
	if parsedArgs.bindPort > 0 && parsedArgs.webHook == "" {
		flag.PrintDefaults()
		log.Fatal("If you specify a port, you have to specify a http hook !")
	}
	if parsedArgs.bindPort < 0 {
		flag.PrintDefaults()
		log.Fatal("bindPort can't be negative!")
	}

	return parsedArgs
}
