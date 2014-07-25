package psdock

import (
	"io"
	"os"
	"os/exec"
)

func redirectIO(cmd *exec.Cmd, f *os.File) {
	var w io.Writer = os.Stdout
	io.Copy(w, f)
}
