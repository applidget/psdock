package psdock

import (
	"io"
	"log"
	"os"
	"testing"
)

func TestParseToml(t *testing.T) {
	tomlFilename := "config_test.toml"
	file, err := os.Create(tomlFilename)
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(tomlFilename)
	tomlConf := `
		Command = "mycommand -v --flag value"
		Stdout = "file:///home/vagrant/output"
		LogRotation = "hourly"
		LogPrefix = "[PRFX]"
		LogColor = "yellow"
		EnvVars = "MYKEY=myval"
		BindPort = 9999
		WebHook = "http://www.myhook.com"
		Stdin = "tcp://stdinServer:1337"
		`
	_, err = io.WriteString(file, tomlConf)
	if err != nil {
		t.Error("Can't create file:" + err.Error())
	}
	conf := Config{}
	parseTOML(&conf, tomlFilename)
	expectedResult := Config{Command: "mycommand -v --flag value", Args: "", Stdout: "file:///home/vagrant/output", LogRotation: "hourly", LogColor: "yellow",
		LogPrefix: "[PRFX]", EnvVars: "MYKEY=myval", BindPort: 9999, WebHook: "http://www.myhook.com", Stdin: "tcp://stdinServer:1337"}
	if conf != expectedResult {
		t.Errorf("expected:%#v\n-got:%#v", expectedResult, conf)
	}
}
