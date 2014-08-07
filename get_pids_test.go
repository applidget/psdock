package psdock

import (
	"strconv"
	"strings"
	"testing"
)

func TestTreeTraversal(t *testing.T) {
	tree := make(map[int][]int)
	tree[100] = []int{200, 500}
	tree[500] = []int{600, 900}
	tree[200] = []int{300}

	myFunctor := func(pid int) string {
		return helper(tree[pid])
	}

	answer, err := treeTraversal(100, myFunctor)
	if err != nil {
		t.Error("err:" + err.Error())
	}
	if helper(answer) != "100\n200\n300\n500\n600\n900" {
		t.Error("result:" + helper(answer) + "-expected 100\n200\n300\n500\n600\n900")
	}
}

func helper(slice []int) string {
	result := []string{}
	for _, elem := range slice {
		result = append(result, strconv.Itoa(elem))
	}
	return strings.Join(result, "\n")
}
