package psdock

import (
	"bufio"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func getPIDs(pid int) ([]int, error) {
	pgrepCmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	pgrepOutput, _ := pgrepCmd.Output()

	scanner := bufio.NewScanner(strings.NewReader(string(pgrepOutput)))
	pids := []int{pid}
	for scanner.Scan() {
		childPid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Println(err)
			break
		}
		childPidsRecur, err := getPIDs(childPid)
		if err != nil {
			log.Println(err)
			break
		}
		pids = append(pids, childPidsRecur...)
	}
	return pids, nil
}
