package psdock

import (
	"errors"
	"os/user"
	"strconv"
	"syscall"
)

//SetUser tries to change the current user to newUsername
func SetUser(newUsername string) error {
	if newUsername == "" {
		return nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return errors.New("Can't determine the current user !\n" + err.Error())
	}

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

	newUserGID, err := strconv.Atoi(newUser.Gid)
	if err != nil {
		return errors.New("Can't determine the new user GID !\n" + err.Error())
	}

	if err := syscall.Setuid(newUserUID); err != nil {
		return errors.New("Can't change the UID to " + newUsername + "!\n" + err.Error())
	}

	if err := syscall.Setgid(newUserGID); err != nil {
		return errors.New("Can't change the GID !\n" + err.Error())
	}
	return nil
}
