package psdock

import (
	"bufio"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//fileLogger is a type used to write the stdout to a file.
type fileLogger struct {
	log          *Logger
	previousName string
	logRotation  string
	filename     string
}

func NewFileLogger(fName, prfx, lRotation string, statusChannel chan ProcessStatus) (*fileLogger, error) {
	result := fileLogger{filename: fName, log: &Logger{prefix: prfx}, logRotation: lRotation}
	err := result.openFirstOutputFile()
	if err != nil {
		return nil, err
	}
	go result.manageLogRotation(statusChannel)
	return &result, nil
}

//retrieveFilenames returns the list of the .log files in the directory containing filename
func retrieveLogFilenames(filename string) ([]string, error) {
	dirName := filepath.Dir(filename)
	fls, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	filenames := []string{}

	//Construct the list of log filenames
	for _, fileinfo := range fls {
		if !fileinfo.IsDir() && filepath.Ext(fileinfo.Name()) == ".log" {
			filenames = append(filenames, dirName+"/"+fileinfo.Name())
		}
	}
	return filenames, nil
}

//constructNewLogFilenames constructs the list of possible new log filenames (which may already exist)
func constructNewLogFilenames(filenames []string, tNow time.Time, lifetime time.Duration, currFilename string) []string {
	result := []string{}
	for _, filename := range filenames {
		if recentEnough(filename, tNow, lifetime) {
			result = append(result, filename)
		}
	}
	return append(result, currFilename+"."+string(tNow.Format("2006-01-02-15-04")+".log"))
}

//openFirstOutputFile opens a file in in which the child's stdout will be written
func (flg *fileLogger) openFirstOutputFile() error {
	//We have to check if one of the files is a log whose start date is less than time.Now()-lifetime.
	//If that's the case, we use that file
	lifetime := convertLogRToDuration(flg.logRotation)

	var f *os.File
	var fName string

	tNow := time.Now()
	filenames, err := retrieveLogFilenames(flg.filename)
	if err != nil {
		return err
	}

	//Construct the list of possible new log filenames
	possibleLogFilenames := constructNewLogFilenames(filenames, tNow, lifetime, flg.filename)
	for _, fName = range possibleLogFilenames {
		f, err = os.OpenFile(fName, os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			//We don't return here since we can try to open other files
			log.Println(err)
		} else {
			break
		}
	}
	flg.log.output = f
	flg.previousName = fName
	return nil
}

//manageLogRotation opens a	new logfile when the lifetime of the current log file is reached.
//The old file is gzipped
func (flg *fileLogger) manageLogRotation(statusChannel chan ProcessStatus) {
	var newName string
	var f *os.File
	lifetime := convertLogRToDuration(flg.logRotation)
	ticker := time.NewTicker(lifetime)
	for {
		_ = <-ticker.C
		//Open the new stdout file
		filenames, err := retrieveLogFilenames(flg.filename)
		if err != nil {
			statusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(constructNewLogFilenames(filenames, time.Now(), lifetime, flg.previousName))))
		if err != nil {
			statusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		for _, newFilename := range filenames {
			f, err = os.Create(newFilename)
			if err != nil {
				statusChannel <- ProcessStatus{Status: -1, Err: err}
			}
		}

		oldOutput := flg.log.output

		//assign it to p.output
		flg.log.output = f
		//Close the previous file
		if err = oldOutput.Close(); err != nil {
			statusChannel <- ProcessStatus{Status: -1, Err: err}
		}

		//gzip&delete previousName
		if err := compressOldOutput(flg.previousName); err != nil {
			statusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		//Save the new name
		flg.previousName = newName
	}
}

//compressOldOutput creates a gzip archive oldFile.gz and puts oldFilePath in it
//it then removes oldFile
func compressOldOutput(oldFile string) error {
	file, err := os.Create(oldFile + ".gz")
	if err != nil {
		return err
	}
	defer file.Close()
	gWriter := gzip.NewWriter(file)
	defer gWriter.Close()
	oldFileOpened, err := os.Open(oldFile)
	if err != nil {
		return err
	}
	bufReader := bufio.NewReader(oldFileOpened)
	if _, err = bufReader.WriteTo(gWriter); err != nil {
		return err
	}
	if err = gWriter.Flush(); err != nil {
		return err
	}
	if err = os.Remove(oldFile); err != nil {
		return err
	}
	return nil
}

//recentEnough tests if the log file path has a timestamp which is nearer to tNow than lifetime
func recentEnough(path string, tNow time.Time, lifetime time.Duration) bool {
	lIndex := strings.LastIndex(path[:len(path)-4], ".") //get rid of the .log extension
	strDateOfPath := path[lIndex+1 : len(path)-4]
	dateOfPath, _ := time.Parse("2006-01-02-15-04", strDateOfPath)
	if tNow.Sub(dateOfPath) < lifetime {
		return true
	} else {
		return false
	}
}

func convertLogRToDuration(lifetime string) time.Duration {
	switch lifetime {
	case "minutely":
		return time.Minute
	case "hourly":
		return time.Hour
	case "daily":
		return time.Hour * 24
	case "weekly":
		return time.Hour * 24 * 7
	}
	return -1
}
