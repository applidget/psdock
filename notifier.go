package psdock

import (
	"bytes"
	"errors"
	"net/http"
)

type Notifier struct {
	webHook string
}

//Notify sends a PUT request to the hook in order to trigger it
func (n Notifier) Notify(status int) error {
	if n.webHook == "" {
		return nil
	}
	statusStr := ""
	if status == PROCESS_STARTED {
		statusStr = "\"starting\""
	} else if status == PROCESS_RUNNING {
		statusStr = "\"up\""
	} else {
		statusStr = "\"crashed\""
	}
	body := `{
							"ps":
								{ "status":` + statusStr + `}
						}`

	req, err := http.NewRequest("PUT", n.webHook, bytes.NewBufferString(body))
	if err != nil {
		return errors.New("Error in Notify : Failed to construct the HTTP request" + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Error in Notify : Was not able to trigger the hook!\n" + err.Error())
	}
	defer resp.Body.Close()

	return nil
}
