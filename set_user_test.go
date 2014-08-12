package psdock

import (
	"os/exec"
	"testing"
)

func TestSetUser(t *testing.T) {
	//the user testUser has to exist
	SetUser("test")
	out, err := exec.Command("whoami").Output()
	if err != nil {
		t.Error(err)
	}
	if string(out) != "test\n" {
		t.Error("The user should be test, not " + string(out))
	}
}
