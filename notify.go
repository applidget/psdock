package psdock

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
)

//NotifyWebHook sends a http "PUT" request to hook. The message is of type json, and
//is "{"ps":{"status":status}}
func NotifyWebHook(hook string, status string) error {
	if hook == "" {
		return nil
	}
	requestMessage := strings.Join([]string{`{"ps":{"status":`, status, `}}`}, "")
	request, err := http.NewRequest("PUT", hook, bytes.NewBufferString(requestMessage))
	if err != nil {
		return errors.New("Failed to contruct the HTTP request.\n" + err.Error())
	}
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	//Send the request
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		return errors.New("Was not able to trigger the hook!\n" + err.Error())
	}
	return nil
}
