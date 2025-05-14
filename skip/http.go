package skip

import (
	"bytes"
	"net/http"
)

func httpPost(url string, body *bytes.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("app-token", AppToken)

	return client.Do(req)
}
