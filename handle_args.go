package psdock

import (
	"errors"
	"flag"
	"fmt"
	"os/user"
	"strings"
)

//command is the name of the command to be executed by psdock
//Config contains the result of command-line-Config parsing.
//args are the Config passed to this command
//stdout is the redirection path for the stdout/stderr of the launched process
//logRotation is the lifetime (in seconds) of a single log file
//logPrefix is the prefix for logging the output of the launched process
//bindPort is the port watched for binding by psdock
//webHook is the hook triggered by psdock in case of special events
//user is the UID of the user launching the process
type Config struct {
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

//ParseConfig parses command-line Config and returns them in an Config struct
func ParseConfig() (*Config, error) {
	parsedConfig := Config{}

	flag.StringVar(&parsedConfig.Command, "process", "", "process to be executed by psdock")
	flag.StringVar(&parsedConfig.Stdout, "stdout", "os.Stdout", "redirection path for the stdout/stderr of the launched process")
	flag.StringVar(&parsedConfig.LogRotation, "log-rotation", "daily", "lifetime of a single log file.")
	flag.StringVar(&parsedConfig.LogPrefix, "log-prefix", "", "prefix for logging the output of the launched process")
	flag.StringVar(&parsedConfig.EnvVars, "env-vars", "", "Config passed to the launched command")
	flag.IntVar(&parsedConfig.BindPort, "bind-port", 0, "port to be watched for binding by psdock(0 means no port is monitored)")
	flag.StringVar(&parsedConfig.WebHook, "web-hook", "", "hook triggered by psdock in case of special events")

	//Retrieve the name of the current user. Will be used as a default value for user-name
	user, err := user.Current()
	if err != nil {
		return nil, errors.New("Failed to retrieve the informations about the current user!\n" + err.Error())
	}
	flag.StringVar(&parsedConfig.UserName, "user-name", user.Username, "name of the user launching the process")

	flag.Parse()
	//The user has to specify a process to run
	if parsedConfig.Command == "" {
		flag.PrintDefaults()
		return nil, errors.New("You must specify a process to run")
	}

	//Split the command given in process name & Config
	commandSplited := strings.SplitAfterN(parsedConfig.Command, " ", 2)
	if len(commandSplited) == 1 {
		parsedConfig.Command = commandSplited[0]
		parsedConfig.Args = ""
	} else {
		parsedConfig.Command = commandSplited[0][:len(commandSplited[0])-1] //drop the last char (' ')
		parsedConfig.Args = commandSplited[1]
	}
	fmt.Println(parsedConfig.Args)

	if parsedConfig.LogRotation != "minutely" && parsedConfig.LogRotation != "hourly" &&
		parsedConfig.LogRotation != "daily" && parsedConfig.LogRotation != "weekly" {
		flag.PrintDefaults()
		return nil, errors.New("logRotation has to be minutely, hourly, daily or weekly !")
	}
	if parsedConfig.BindPort > 0 && parsedConfig.WebHook == "" {
		flag.PrintDefaults()
		return nil, errors.New("If you specify a port, you have to specify a http hook !")
	}
	if parsedConfig.BindPort < 0 {
		flag.PrintDefaults()
		return nil, errors.New("bindPort can't be negative!")
	}

	return &parsedConfig, nil
}
