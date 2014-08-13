**psdock**
======

[![Build Status](https://travis-ci.org/applidget/psdock.svg)](https://travis-ci.org/applidget/psdock)

A simple tool to launch and monitor processes.

#Installation


1) Make sure $GOPATH/bin is in your path and install godep  
`````bash
$ go get github.com/kr/godep  
$ export PATH=$PATH:$GOPATH/bin
`````
2) Get psdock and install it
`````
$ go get github.com/applidget/psdock  
$ cd $GOPATH/src/github.com/applidget/psdock  
$ make
`````

#Usage

###Basic
Psdock can launch a process very simply :

    psdock --command ls
###Configuration files
Configuration files can be specified in this way :

    psdock -c config.toml
    
If no argument is given, psdock will automatically search for the file at `/etc/psdock/psdock.conf`.

Here is an example of config.toml :
    `````toml
    Command = "nc -l 8080"
    Webhook = "http://distantUrl:80"
    Bindport = 8080
    Logprefix= "NETCAT"
    `````
    
###Stdout
Three types of output can be specified :
* Standard output : 
    This is the output used if the -stdout option is not specified. The output from the process will be written on the standard output.
* Logfile : 
    For instance, you tell psdock to write the output of the process to a file named `bashLog`. Log-rotation is then automatically handled : by default, log files are rotated every day, but you can tell to psdock to rotate logs every minute, every hour, or every week :
    
        psdock --command bash --stdout "file:///bashLog" --log-rotation "hourly"
    Instead of writing directly to `bashLog`, psdock will write the output to `bashLog.YYYY-MM-DD-hh-mm.log`, where the date is the creation time of this file. When its lifetime will expire, psdock will compress this file to a gzip archive named `bashLog.YYYY-MM-DD-hh-mm.tar.gz` and will start writing the output to a new file. Psdock also ensures that the number of archived logs will not exceed 5.
* TCP Socket : 
    The output of the process can be send through a TCP connection :

        psdock --command bash --stdout "tcp://localhost:666"

###Stdin
Two types of stdin can be specified :
* Standard input : 
    This is the input used if the -stdin option is not specified. The data read on the standard input will be passed to the process.
* TCP Socket : 
    The input of the process can be read from a TCP socket : 

        psdock --command bash --stdin "tcp://localhost:666"

###Log Formatting
You can specify a prefix for the output of the process, and set its color : 

    psdock --command ls --log-prefix "[PREFIX]" --log-color "red"

The color can be `"black"` (default), `"white"`, `"red"`, `"green"`, `"blue"`, `"yellow"`, `"cyan"` or `"magenta"`. 

###Web Hook
A web hook can be specified as a flag:

    psdock --command "bash" --web-hook "http://distantServer:3000"
Psdock will send status informations about the process to the web hook, through a PUT HTTP request. The body will be formatted like this :

    {ps: { status: stat}}

where `stat` can be `PROCESS_STARTED`, `PROCESS_RUNNING` or `PROCESS_STOPPED`.

###BindPort
If the --bind-port flag is set then psdock will wait for the process to bind the port specified before sending the `PROCESS_RUNNING` message to the web hook : 

    psdock --command "nc -l 8080" --web-hook "http://distantUrl:3000" --bind-port 8080

###SetUser
The process can be executed under a different user. To do so, you have to specify the username of the desired process owner : 

    psdock --command bash --set-user "alice"

###Environment variables
You can specify the environment variables to set in the process execution context : 

    psdock --command bash --env-vars "LD_PRELOAD=\"/path/to/my/malloc.so\""
    
#License

Psdock is licensed under the MIT license. See LICENSE for the full text.
