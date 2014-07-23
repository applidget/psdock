package psdock

import (
	"os"
	"os/exec"
	"strings"
)

//SetEnvVars sets the environment variables for the launched process
func SetEnvVars(c *exec.Cmd, envVars string) {
	//We first have to manually copy all the current environment variables to c.Env
	c.Env = append(c.Env, os.Environ()...)
	for _, str := range strings.Split(envVars, " ") {
		c.Env = append(c.Env, str)
	}
}
