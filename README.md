**psdock**
======

A simple tool to launch and monitor processes. 

**Usage :**

Example : 

  `psdock --process "nc -l 8080" --web-hook http://distantUrl:80 --bind-port 8080 --log-prefix "NETCAT"`
  
Flags : 
  * --process : process to be executed by psdock
  * --stdout : redirection path for the stdout/stderr of the launched process (stdout by default)
  * --log-rotation : lifetime of a single log file. Can be "hourly", "daily" (default), "monthly" or "weekly"
  * --log-prefix : prefix for logging the output of the launched process
  * --env-vars : arguments passed to the launched command. They have to be passed as *"KEY1=value1 KEY2=value2"*. 
  * --bind-port : port to be watched for binding by psdock
  * --web-hook : hook triggered by psdock in case of special events

