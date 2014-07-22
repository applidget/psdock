package psdock

import (
	"exec"
	"os"
	"strings"
)

//SetEnvVars sets the environment variables for the launched process
func (c *Cmd) SetEnvVars(envVars string) {
	if envVars == "" {
		//We don't want to export any environment variable
		//But we can't leave c.Env empty, as exec.Run will export all the current
		//env vars. So we only export one unused environment variable
		c.Env.append("UNUSED")
	}
	for _, str := range strings.Split(envVars, " ") {
		//We want to export all the psdock's environment variables to the launched
		//process' environment
		if str == "$EXPORTALL" {
			if len(envVars) == 1 {
				//We want to export all the current environment variables, but nothing else.
				//We can therefore leave envVars empty.
				c.Env = make([]string, 0)
			} else {
				//We have new environment variables to export.
				//We have to manually copy all the current environment variables to c.Env
				c.Env = append(c.Env, os.Environ)
			}
			//We want to export a specific variable from the current environment
		} else if str[0] == '$' {
			valueOfEnv := os.Getenv(str[1:])
			if valueOfEnv == "" {
				log.Fatal("The variable", str[1:], "is not a defined environment variable")
			} else {
				c.Env.append([]string{str[1:], "=", valueOfEnv})
			}
			//str is already in the form "VAR1=value1"
		} else {
			c.Env.append(str)
		}
	}
}
