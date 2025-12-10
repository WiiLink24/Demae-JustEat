package skip

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func httpPost(url string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("app-token", AppToken)

	return client.Do(req)
}
