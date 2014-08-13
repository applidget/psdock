**psdock**
======

[![Build Status](https://travis-ci.org/applidget/psdock.svg)](https://travis-ci.org/applidget/psdock)

A simple tool to launch and monitor processes.

#Installation
------------

1) Make sure $GOPATH/bin is in your path and install godep  
`go get github.com/kr/godep`  
`export PATH=$PATH:$GOPATH/bin`  
2) Get psdock and install it  
`go get github.com/applidget/psdock`  
`cd $GOPATH/src/github.com/applidget/psdock`  
`make`


#Usage
------------
###Basic
Ps-dock can launch a process very simply, in this way:

    `psdock --command ls`
###Config
Config file can be specified in this way :

    `psdock -c config.toml`
    
If no argument is given, ps-dock will automatically search for the file at /etc/psdock/psdock.conf.

Here is an example of .psdockrc :
    `````toml
    Command = "nc -l 8080"
    Webhook = "http://distantUrl:80"
    Bindport = 8080
    Logprefix= "NETCAT"
    `````
    
###Stdout
Three types of stdout can be specified :
* Standard output
    This is the stdout used if the -stdout option is not specified. The output from the process will be written on the standard output.
* Logfile

    For instance, you can specify a file name test.log to ps-dock. Log-rotation is the automatically handled : by defaults, log files are rotated every day, but you can tell to ps-dock to rotate logs every minute, every hour, or every week in this way:
    
        `psdock --command "bash" --stdout "file:///test.log" --log-rotation "hourly"`

* TCP Socket
    A distant socket to which send datas from process.

        `psdock --command "bash" --stdout "tcp://localhost:666"`

###Stdin
Two types of stdin can be specified :
* Standard input
    This is the stdin used if the -stdin option is not specified. The data read on the standard input will be passed to the process
* TCP Socket
    A distant socket to which read data to pass to the process.

        `psdock --command "bash" --stdin "tcp://localhost:666"`

###Log Formatting
You can specify a prefix for the output of the process, and set its color : 

    `psdock --command ls --log-prefix "[PREFIX]" --log-color "red"`

The color can be `"black"` (default), `"white"`, `"red"`, `"green"`, `"blue"`, `"yellow"`, `"cyan"` or `"magenta"`. 

###Web Hook
A web hook can be specified as a flag:

    `psdock --command "bash" --web-hook "http://distantServer:3000"`
Or in config file as specified before.
It will send status informations about the process launched by ps-dock to the web hook. Body of datas sent are formatted likethis :

    `{ps: { status: stat}}`

where stat can be PROCESS_STARTED, PROCESS_RUNNING or PROCESS_STOPPED.

###BindPort
If the --bind-port flag is set then psdock will wait that process open port specified in environment variable to inform Web Hook that process status is up.

    psdock --command "nc -l 8080" --web-hook "http://distantUrl:3000" --bind-port 8080

###SetUser
The process can be executed under a different user. To do so, you have to specify the username of the desired process owner : 

    psdock --command bash --set-user "alice"


#License
------------
Psdock is licensed under the MIT license. See LICENSE for the full text.
