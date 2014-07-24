package psdock

import (
	"errors"
	"flag"
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
type Arguments struct {
	Command     string
	Args        string
	Stdout      string
	LogRotation string
	LogPrefix   string
	EnvVars     string
	BindPort    int
	WebHook     string
	UserName    string
}

//ParseArguments parses command-line arguments and returns them in an arguments struct
func ParseArguments() (*Arguments, error) {
	parsedArgs := Arguments{}

	flag.StringVar(&parsedArgs.Command, "process", "", "process to be executed by psdock")
	flag.StringVar(&parsedArgs.Stdout, "stdout", "os.Stdout", "redirection path for the stdout/stderr of the launched process")
	flag.StringVar(&parsedArgs.LogRotation, "log-rotation", "daily", "lifetime of a single log file.")
	flag.StringVar(&parsedArgs.LogPrefix, "log-prefix", "", "prefix for logging the output of the launched process")
	flag.StringVar(&parsedArgs.EnvVars, "env-vars", "", "arguments passed to the launched command")
	flag.IntVar(&parsedArgs.BindPort, "bind-port", 0, "port to be watched for binding by psdock(0 means no port is monitored)")
	flag.StringVar(&parsedArgs.WebHook, "web-hook", "", "hook triggered by psdock in case of special events")

	//Retrieve the name of the current user. Will be used as a default value for user-name
	user, err := user.Current()
	if err != nil {
		return nil, errors.New("Failed to retrieve the informations about the current user!\n" + err.Error())
	}
	flag.StringVar(&parsedArgs.UserName, "user-name", user.Username, "name of the user launching the process")

	flag.Parse()
	//The user has to specify a process to run
	if parsedArgs.Command == "" {
		flag.PrintDefaults()
		return nil, errors.New("You must specify a process to run")
	}

	//Split the command given in process name & arguments
	commandSplited := strings.SplitAfterN(parsedArgs.Command, " ", 2)
	parsedArgs.Command = commandSplited[0][:len(commandSplited[0])-1] //drop the last char (' ')
	if len(commandSplited) == 1 {
		parsedArgs.Args = ""
	} else {
		parsedArgs.Args = commandSplited[1]
	}

	if parsedArgs.LogRotation != "minutely" && parsedArgs.LogRotation != "hourly" &&
		parsedArgs.LogRotation != "daily" && parsedArgs.LogRotation != "weekly" {
		flag.PrintDefaults()
		return nil, errors.New("logRotation has to be minutely, hourly, daily or weekly !")
	}
	if parsedArgs.BindPort > 0 && parsedArgs.WebHook == "" {
		flag.PrintDefaults()
		return nil, errors.New("If you specify a port, you have to specify a http hook !")
	}
	if parsedArgs.BindPort < 0 {
		flag.PrintDefaults()
		return nil, errors.New("bindPort can't be negative!")
	}

	return &parsedArgs, nil
}
