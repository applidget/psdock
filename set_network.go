package psdock

import (
	"errors"
	"log"
	"os/exec"
)

func SetNetwork() error {
	cmd := exec.Command("ip", "route", "add", "default", "via", "10.0.3.1")
	out, err := cmd.Output()
	if err != nil {
		log.Println(out)
		return errors.New("Can't set gw:" + err.Error())
	}
	return nil
}
