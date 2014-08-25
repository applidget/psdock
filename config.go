package psdock

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"os"
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
	LogColor    string
	EnvVars     string
	BindPort    int
	WebHook     string
	Stdin       string
	UserName    string
	Gateway     string
}

//ParseArgs parses command-line Config and returns them in an Config struct
func ParseArgs() (*Config, error) {
	parsedConfig := Config{}
	var tomlConfigFilename string

	flag.StringVar(&parsedConfig.Command, "command", "", "command to be executed by psdock")
	flag.StringVar(&parsedConfig.Stdout, "stdout", "os.Stdout", "redirection path for the stdout/stderr of the launched process")
	flag.StringVar(&parsedConfig.LogRotation, "log-rotation", "daily", "lifetime of a single log file.")
	flag.StringVar(&parsedConfig.LogPrefix, "log-prefix", "", "prefix for logging the output of the launched process")
	flag.StringVar(&parsedConfig.LogColor, "log-prefix-color", "black", "color for logging the output of the launched process")
	flag.StringVar(&parsedConfig.EnvVars, "env-vars", "", "Config passed to the launched command")
	flag.IntVar(&parsedConfig.BindPort, "bind-port", 0, "port to be watched for binding by psdock (0 means no port is monitored)")
	flag.StringVar(&parsedConfig.WebHook, "web-hook", "", "hook triggered by psdock in case of special events")
	flag.StringVar(&parsedConfig.Stdin, "stdin", "os.Stdin", "url used to read stdin")
	flag.StringVar(&tomlConfigFilename, "c", "", "filename of the toml file used to read the config")

	//Retrieve the name of the current user. Will be used as a default value for user-name
	user, err := user.Current()
	if err != nil {
		return nil, errors.New("Error in ParseArgs():Failed to retrieve the informations about the current user!\n" + err.Error())
	}
	flag.StringVar(&parsedConfig.UserName, "user-name", user.Username, "name of the user launching the process")

	flag.Parse()

	if tomlConfigFilename != "" {
		if len(os.Args) != 3 {
			flag.PrintDefaults()
			return nil, errors.New("Error in ParseArgs():If a toml config file is given, you can't specify other arguments!")
		}
		err := parseTOML(&parsedConfig, tomlConfigFilename)
		if err != nil {
			return nil, errors.New("Error in ParseArgs():Can't parse TOML file:" + err.Error())
		}
	}

	//We haven't specified a file with the -c command. Let's still try to open PSDOCK_CFG_FILEPATH
	if len(os.Args) == 1 {
		if _, err = os.Stat(PSDOCK_CFG_FILEPATH); err == nil {
			err = parseTOML(&parsedConfig, PSDOCK_CFG_FILEPATH)
			if err != nil {
				return nil, errors.New("Error in ParseArgs():Can't parse TOML file:" + err.Error())
			}
		}
	}
	//The user has to specify a process to run
	if parsedConfig.Command == "" {
		flag.PrintDefaults()
		return nil, errors.New("Error in ParseArgs():You must specify a process to run")
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

	if parsedConfig.LogRotation != "minutely" && parsedConfig.LogRotation != "hourly" &&
		parsedConfig.LogRotation != "daily" && parsedConfig.LogRotation != "weekly" {
		flag.PrintDefaults()
		return nil, errors.New("Error in ParseArgs():log-rotation has to be minutely, hourly, daily or weekly !")
	}

	if parsedConfig.LogColor != "black" && parsedConfig.LogColor != "white" &&
		parsedConfig.LogColor != "red" && parsedConfig.LogColor != "green" &&
		parsedConfig.LogColor != "blue" && parsedConfig.LogColor != "yellow" &&
		parsedConfig.LogColor != "magenta" && parsedConfig.LogColor != "cyan" {
		flag.PrintDefaults()
		return nil, errors.New("Error in ParseArgs():log-prefix-color has to be black, white, red, green, blue, yellow, cyan or magenta !")
	}

	if parsedConfig.BindPort > 0 && parsedConfig.WebHook == "" {
		flag.PrintDefaults()
		return nil, errors.New("Error in ParseArgs():If you specify a port, you have to specify a http hook !")
	}
	if parsedConfig.BindPort < 0 {
		flag.PrintDefaults()
		return nil, errors.New("Error in ParseArgs():bind-port can't be negative!")
	}

	return &parsedConfig, nil
}

//parseTOML parses a toml file and fills conf
func parseTOML(conf *Config, filename string) error {
	if _, err := toml.DecodeFile(filename, conf); err != nil {
		return err
	}
	return nil
}
