package psdock

import (
	"testing"
)

func TestHasBoundPort(t *testing.T) {
	f := func(p int) ([]int, error) {
		if p == 5910 {
			return []int{5910}, nil
		} else {
			return []int{}, nil
		}
	}
	lsofOutput := `COMMAND  PID        USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
nc    5910 developpeur    3u  IPv4 0xae0388cd1b51214d      0t0  TCP *:http-alt (LISTEN)`
	if !parseLsof([]byte(lsofOutput), 5910, f) {
		t.Error("Should return true")
	}
}
