package psdock

import (
	"os"
	"os/exec"
	"strings"
)

//SetEnvVars sets the environment variables for the launched process
//Note:if we precise a variable already present in the user's environment,
//the value will be the one we have given.
//The $PATH variable is not empty by default. If we override it, the elements we
//provide will be appended to the default $PATH (/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin)
func SetEnvVars(c *exec.Cmd, envVars string) {
	//We first have to manually copy all the current environment variables to c.Env
	c.Env = append(c.Env, os.Environ()...)
	for _, str := range strings.Split(envVars, " ") {
		c.Env = append(c.Env, str)
	}
}
