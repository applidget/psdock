package main

import (
	//"github.com/applidget/psdock"
	"os/exec"
	"testing"
)

func TestPsdock(t *testing.T) {
	cmd := exec.Command("/home/vagrant/code/go/bin/psdock")
	cmd2 := exec.Command("ls")
	out, err := cmd.Output()
	if err != nil {
		t.Error("psdock exec error:" + err.Error())
	}
	out2, err := cmd2.Output()
	if err != nil {
		t.Error("ls exec error:" + err.Error())
	}
	if string(out) != string(out2) {
		t.Error("expected:" + string(out) + "-got:" + string(out2))
	}
}
