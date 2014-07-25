package psdock

import (
	"errors"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

//PrepareProcess sets the environment variables & changes the user
func PrepareProcessEnv(cmd *exec.Cmd, Config *Config) error {
	setEnvVars(cmd, Config.EnvVars)
	if err := changeUser(Config.UserName); err != nil {
		return err
	}
	return nil
}

//setEnvVars sets the environment variables for the launched process
//Note:if we precise a variable already present in the user's environment,
//the value will be the one we have given.
//The $PATH variable is not empty by default. If we override it, the elements we
//provide will be appended to the existing $PATH
func setEnvVars(c *exec.Cmd, envVars string) {
	//We first have to manually copy all the current environment variables to c.Env
	if len(envVars) == 0 {
		return
	}
	//Uncomment the line below to automatically export the current env vars to the
	//child process context
	//c.Env = append(c.Env, os.Environ()...)
	for _, str := range strings.Split(envVars, " ") {
		c.Env = append(c.Env, str)
	}
}

//changeUser tries to change the current user to newUsername.
func changeUser(newUsername string) error {
	currentUser, err := user.Current()
	if err != nil {
		return errors.New("Can't determine the current user !\n" + err.Error())
	}

	//If newUserName is the current username, we return
	if newUsername == currentUser.Username {
		return nil
	}

	newUser, err := user.Lookup(newUsername)
	if err != nil {
		return errors.New("Can't find the user" + newUsername + "!\n" + err.Error())
	}

	newUserUID, err := strconv.Atoi(newUser.Uid)
	if err != nil {
		return errors.New("Can't determine the new user UID !\n" + err.Error())
	}

	if err := syscall.Setuid(newUserUID); err != nil {
		return errors.New("Can't change the user !\n" + err.Error())
	}
	return nil
}
