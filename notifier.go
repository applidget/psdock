package psdock

import (
	"bytes"
	"errors"
	"net/http"
)

type Notifier struct {
	webHook string
}

func (n Notifier) NotifyStatusChanged(status int) error {
	if n.webHook == "" {
		return nil
	}
	statusStr := ""
	if status == PROCESS_STARTED {
		statusStr = "started"
	} else if status == PROCESS_RUNNING {
		statusStr = "running"
	} else {
		statusStr = "stopped"
	}
	body := `{
							"ps":
								{ "status":` + statusStr + `}
						}`

	req, err := http.NewRequest("PUT", n.webHook, bytes.NewBufferString(body))
	if err != nil {
		return errors.New("Failed to construct the HTTP request" + err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return errors.New("Was not able to trigger the hook!\n" + err.Error())
	}
	return nil
}
