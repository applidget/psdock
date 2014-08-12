**psdock**
======

[![Build Status](https://travis-ci.org/applidget/psdock.svg)](https://travis-ci.org/applidget/psdock)

A simple tool to launch and monitor processes.

**Install :**


1) Make sure $GOPATH/bin is in your path and install godep  
`go get github.com/kr/godep`  
`export PATH=$PATH:$GOPATH/bin`  
2) Get psdock and install it  
`go get github.com/applidget/psdock`  
`cd $GOPATH/src/github.com/applidget/psdock`  
`make`


**Usage :**

Example :

  `````
  psdock --command "nc -l 8080" --web-hook "http://distantUrl:80" --bind-port 8080 --log-prefix "NETCAT"
  `````

Flags :  
  * `--command` : command to be executed by psdock  
  * `--stdout` : redirection path for the stdout/stderr of the launched process (`"os.Stdout"` by default). Can alse be a file (`"file:///home/ubuntu/logs/netcat.log"`) or a TCP socket (`"tcp://server:1337"`)
  * `--log-rotation` : lifetime of a single log file. Can be `"minutely"`, `"daily"` (default), `"hourly"` or `"weekly"`
  * `--log-prefix` : prefix for logging the output of the launched process
  * `--log-color` : color of the prefix. Can be `"black"` (default), `"white"`, `"red"`, `"green"`, `"blue"`, `"yellow"`, `"cyan"` or `"magenta"`
  * `--env-vars` : arguments passed to the launched command. They have to be passed as `"KEY1=value1 KEY2=value2"`.  
  * `--bind-port` : port to be watched for binding by psdock  
  * `--web-hook` : hook triggered by psdock in case of special events. Has to be a http URL.  
  * `--stdin` : path used to read the stdin passed to the launched process. Can be `"os.Stdin"` (default) or a TCP socket
  * `-c` : filepath of the TOML file used to read the arguments. No other flag can be passed if the -c flag is used

  
TOML config files must have the follow the standard syntax : 
  
  `````toml
  command = "nc -l 8080"
  webhook = "http://distantUrl:80"
  bindport = 8080
  logprefix= "NETCAT"
  `````
 
If no argument is given to psdock, it will attempt to read the configuration from `/etc/psdock/psdock.conf` 
  
