package psdock

import (
	"bufio"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

//openFirstOutputFile tries to open a file in order to redirect stdout. If a
func (flg *fileLogger) openFirstOutputFile() error {
	//We have to check if one of the files is a log whose start date is less than time.Now()-lifetime.
	//If that's the case, we use that file
	lifetime := convertLogRToDuration(flg.logRotation)
	dirName := filepath.Dir(flg.filename)
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}
	var f *os.File
	tNow := time.Now()
	for _, fileinfo := range files {
		name := dirName + "/" + fileinfo.Name()
		if fileinfo.IsDir() == false && filepath.Ext(name) == ".log" && recentEnough(fileinfo.Name(), tNow, lifetime) {
			f, err = os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				//We don't return here since we can try to open other files
				log.Println(err)
			} else {
				flg.log.output = f
				flg.previousName = name
				return nil
			}
		}
	}
	//If we arrive here it means we haven't correctly opened a file.
	//We therefore create a new one
	newName := flg.filename + "." + string(time.Now().Format("2006-01-02-15-04")+".log")
	flg.previousName = newName
	f, err = os.Create(newName)
	flg.log.output = f
	return err
}

func (flg *fileLogger) manageLogRotation(statusChannel chan ProcessStatus) {
	var newName string
	lifetime := convertLogRToDuration(flg.logRotation)
	ticker := time.NewTicker(lifetime)
	for {
		_ = <-ticker.C
		//Open the new stdout file
		newName = flg.filename + "." + string(time.Now().Format("2006-01-02-15-04")+".log")
		f, err := os.Create(newName)
		if err != nil {
			statusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		oldOutput := flg.log.output

		//assign it to p.output
		flg.log.output = f
		//we have to close the previous file in order for the copy to be done in the new stdout.
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
