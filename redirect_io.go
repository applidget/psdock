package psdock

import (
	"io"
	"os"
	"os/exec"
)

func redirectIO(cmd *exec.Cmd, f *os.File, stdout string) {
	go io.Copy(f, os.Stdin)
	var w io.Writer

	if stdout == "os.Stdout" {
		w = os.Stdout
	}
	io.Copy(w, f)
}
