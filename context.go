package psdock

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

//SetEnvVars sets the environment variables for the launched process
//Note:if we precise a variable already present in the user's environment,
//the value will be the one we have given.
//The $PATH variable is not empty by default. If we override it, the elements we
//provide will be appended to the default $PATH (/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin)
func SetEnvVars(c *exec.Cmd, envVars string) {
	//We first have to manually copy all the current environment variables to c.Env
	if len(envVars) == 0 {
		return
	}
	c.Env = append(c.Env, os.Environ()...)
	for _, str := range strings.Split(envVars, " ") {
		c.Env = append(c.Env, str)
	}
}

//ChangeUser tries to change the current user to newUsername.
func ChangeUser(newUsername string) error {
	currentUser, err := user.Current()
	if err != nil {
		return errors.New("Can't determine the current user !\n" + err)
	}

	//If newUserName is the current username, we return
	if newUsername == currentUser.Username {
		return nil
	}

	newUser, err := user.Lookup(newUsername)
	if err != nil {
		return errors.New("Can't find the user" + newUser + "!\n" + err)
	}

	newUserUID, err := strconv.Atoi(newUser.Uid)
	if err != nil {
		return errors.New("Can't determine the new user UID !\n" + err)
	}

	if err := syscall.Setuid(newUserUID); err != nil {
		return errors.New("Can't change the user !\n" + err)
	}
	return nil
}
