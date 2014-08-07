package psdock

import (
	"bufio"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

//Returns the list of PIDs : pid + the PIDs of the children of pid (& recursively)
func getPIDs(pid int) ([]int, error) {
	return treeTraversal(pid, retrieveChildrenFromPgrep)
}

//treeTraversal performs a DFS-traversal of functor(tree)
func treeTraversal(pid int, functor func(int) string) ([]int, error) {
	functorOutput := functor(pid)
	scanner := bufio.NewScanner(strings.NewReader(functorOutput))
	result := []int{pid}
	for scanner.Scan() {
		child, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Println(err)
			break
		}
		childRecur, err := treeTraversal(child, functor)
		if err != nil {
			log.Println(err)
			break
		}
		result = append(result, childRecur...)
	}
	return result, nil
}

//retrieveChildrenFromPgrep returns the output of pgrep -P pid
func retrieveChildrenFromPgrep(pid int) string {
	pgrepCmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	pgrepOutput, _ := pgrepCmd.Output()
	return string(pgrepOutput)
}
