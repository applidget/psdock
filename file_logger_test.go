package psdock

import (
	"strings"
	"testing"
	"time"
)

func TestGetLogFilenames(t *testing.T) {
	filenames := []string{"bash.2006-01-02-15-04.log"}
	resultExpected := strings.Join([]string{"bash.2006-01-02-15-04.log", "bash.2006-01-02-15-34.log"}, "")
	result := strings.Join(constructNewLogFilenames(filenames, time.Date(2006, time.January, 2, 15, 34, 0, 0, time.UTC), time.Hour, "bash"), "")
	if result != resultExpected {
		t.Error("Expected:" + resultExpected + "\ngot:" + result)
	}

	resultExpected = strings.Join([]string{"bash.2006-01-02-16-34.log"}, "")
	result = strings.Join(constructNewLogFilenames(filenames, time.Date(2006, time.January, 2, 16, 34, 0, 0, time.UTC), time.Hour, "bash"), "")
	if result != resultExpected {
		t.Error("Expected:" + resultExpected + "\ngot:" + result)
	}
}
