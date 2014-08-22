package psdock

import (
	"errors"
	"log"
	"os/exec"
)

//SetNetwork sets the gateway of the container. It should only be called if psdock is used in "/sbin/init" mode
func SetGateway(gateway string) error {
	cmd := exec.Command("ip", "route", "add", "default", "via", gateway)
	out, err := cmd.Output()
	if err != nil {
		log.Println(out)
		return errors.New("Error in SetGateway:" + err.Error())
	}
	return nil
}
