package psdock

import (
	"os"
	"strconv"
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

func TestConvertRToDuration(t *testing.T) {
	if time.Hour != convertLogRToDuration("hourly") {
		t.Error("Should return time.Hour")
	}
	if time.Minute != convertLogRToDuration("minutely") {
		t.Error("Should return time.Minute")
	}
	if time.Hour*24 != convertLogRToDuration("daily") {
		t.Error("Should return time.Hour*24")
	}
	if time.Hour*24*7 != convertLogRToDuration("weekly") {
		t.Error("Should return time.Hour*24*7")
	}
}

func TestRetrieveFilenames(t *testing.T) {
	extension := strings.Replace(strings.Replace(time.Now().String(), " ", "", -1), ".", "", -1) + "log"
	_, err := os.Create("test." + extension)
	defer os.Remove("test." + extension)
	if err != nil {
		t.Error("Can't create test." + extension + ":" + err.Error())
	}
	fNames, err := retrieveFilenames("test", "."+extension)
	if err != nil {
		t.Error(err)
	}
	if fNames[0][2:] != "test."+extension {
		t.Error("Should be " + "test." + extension + ", got:" + fNames[0])
	}
}

func TestGetFiveLast(t *testing.T) {
	functor := func(s0, s1 string) ([]string, error) {
		return []string{"archive.2000-01-02-16-34.gz", "archive.2001-01-02-16-34.gz", "archive.2002-01-02-16-34.gz", "archive.2003-01-02-16-34.gz",
			"archive.2004-01-02-16-34.gz", "archive.2005-01-02-16-34.gz", "archive.2006-01-02-16-34.gz"}, nil
	}
	fToDelete, err := getFiveLast("archive", 5, functor)
	if err != nil {
		t.Error(err)
	}
	if len(fToDelete) != 2 {
		t.Error("fToDelete should be of length 2 not " + strconv.Itoa(len(fToDelete)))
	}
	if fToDelete[0] != "archive.2005-01-02-16-34.gz" /*|| fToDelete[1] != "archive.2006-01-02-16-34.gz"*/ {
		t.Error("Got:" + fToDelete[0] + "," + fToDelete[1])
		t.Error("Expected:{\"archive.2005-01-02-16-34.gz\", \"archive.2006-01-02-16-34.gz\"}")
	}
}
