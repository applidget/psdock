package psdock

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
)

func redirectIO(cmd *exec.Cmd, f *os.File) {
	go io.Copy(f, os.Stdin)
	scanner := bufio.NewScanner(f)
	//_ = scanner.Scan()
	for scanner.Scan() {
		io.Copy(os.Stdout, strings.NewReader(scanner.Text()))
	}
	//io.Copy(os.Stdout, f)
}
