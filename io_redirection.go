package psdock

import (
	"bufio"
	"compress/gzip"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (p *Process) redirectStdout() error {
	if p.Conf.Stdout != "os.Stdout" {
		url, err := url.Parse(p.Conf.Stdout)
		if err != nil {
			return err
		}
		if url.Scheme == "file" {
			newFilename, err := p.openFirstOutputFile(url.Host + url.Path)
			if err != nil {
				return err
			}
			go p.manageLogRotation(url.Host+url.Path, newFilename)
		} else if url.Scheme == "tcp" {
			p.output, err = net.Dial("tcp", url.Host+url.Path)
			if err != nil {
				p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
			}
		}
	}
	go p.startCopy()
	return nil
}

func (p *Process) startCopy() {
	_, _ = p.output.Write([]byte(p.Conf.LogPrefix))
	reader := bufio.NewReader(p.Pty)
	for {
		rune, _, _ := reader.ReadRune()
		_, err := p.output.Write([]byte{byte(rune)})
		if rune == '\n' {
			_, _ = p.output.Write([]byte(p.Conf.LogPrefix))
		}
		if err != nil {
			break
		}
	}
	//If we arrive here, the logger has created a new file, and it is assigned to p.output
	//We start writing on the new p.output
	p.startCopy()
}

//openFirstOutputFile tries to open a file in order to redirect stdout. If a
func (p *Process) openFirstOutputFile(filename string) (string, error) {
	//We have to check if one of the files is a log whose start date is less than time.Now()-lifetime.
	//If that's the case, we use that file
	lifetime := convertLogRToDuration(p.Conf.LogRotation)
	dirName := filepath.Dir(filename)
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Print(err)
	}
	var f *os.File
	tNow := time.Now()
	for _, fileinfo := range files {
		name := dirName + "/" + fileinfo.Name()
		if fileinfo.IsDir() == false && filepath.Ext(name) == ".log" && recentEnough(fileinfo.Name(), tNow, lifetime) {
			f, err = os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				//We don't return here since we can try to open other files
				log.Print(err.Error())
			} else {
				p.output = f
				return name, nil
			}
		}
	}
	//If we arrive here it means we haven't correctly opened a file.
	//We therefore create a new one
	newName := filename + "." + string(time.Now().Format("2006-01-02-15-04")+".log")
	f, err = os.Create(newName)
	p.output = f
	return newName, err
}

func (p *Process) manageLogRotation(filename, pName string) {
	var newName, previousName string
	previousName = pName
	lifetime := convertLogRToDuration(p.Conf.LogRotation)
	ticker := time.NewTicker(lifetime)
	for {
		_ = <-ticker.C
		//Open the new stdout file
		newName = filename + "." + string(time.Now().Format("2006-01-02-15-04")+".log")
		f, err := os.Create(newName)
		if err != nil {
			p.StatusChannel <- ProcessStatus{Status: -1, Err: err}
		}
		oldOutput := p.output

		//assign it to p.output
		p.output = f

		//we have to close the previous file in order for the copy to be done in the new stdout.
		if err = oldOutput.Close(); err != nil {
			log.Print(err)
		}
		//gzip&delete previousName
		if err := compressOldOutput(previousName); err != nil {
			log.Print("Can't archive old file:" + err.Error())
		}
		//Save the new name
		previousName = newName
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
	fileContent, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return err
	}
	if _, err = gWriter.Write(fileContent); err != nil {
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
