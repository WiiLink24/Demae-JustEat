package justeat

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var UserAgent = func(deviceId string) string {
	return fmt.Sprintf("[JUST-EAT-APP/%s/Android - %s - 11 (API 30)]", ApplicationVersion, deviceId)
}

const (
	ApplicationID      = "4"
	ApplicationVersion = "11.0.0.1610004768"
	Accept             = "application/json,text/json"
	AcceptVersion      = "2"
	JetApplicationID   = "16"
	JetVersion         = "11.0.0.1610004768"
)

func (j *JEClient) httpGet(url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", UserAgent(j.DeviceModel))
	req.Header.Set("Application-Id", ApplicationID)
	req.Header.Set("Application-Version", ApplicationVersion)
	req.Header.Set("Accept-Language", languageCodes[j.Country])
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Accept-Tenant", string(j.Country))
	req.Header.Set("Accept", Accept)
	req.Header.Set("Accept-Version", AcceptVersion)
	req.Header.Set("Authorization", j.Auth)
	req.Header.Set("X-Jet-Application-Id", JetApplicationID)
	req.Header.Set("X-Jet-Application-Version", JetVersion)

	return client.Do(req)
}

func (j *JEClient) httpDO(url string, body any, method string) (*http.Response, error) {
	// First encode everything
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, _ := http.NewRequest(method, url, bytes.NewReader(data))

	req.Header.Set("User-Agent", UserAgent(j.DeviceModel))
	req.Header.Set("Application-Id", ApplicationID)
	req.Header.Set("Application-Version", ApplicationVersion)
	req.Header.Set("Accept-Language", languageCodes[j.Country])
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Accept-Tenant", string(j.Country))
	req.Header.Set("Accept", Accept)
	req.Header.Set("Accept-Version", AcceptVersion)
	req.Header.Set("Authorization", j.Auth)
	if method == http.MethodPatch {
		// The PATCH request requires this specific header.
		req.Header.Set("Content-Type", "application/json; v=2; charset=utf-8")
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	req.Header.Set("X-Jet-Application-Id", JetApplicationID)
	req.Header.Set("X-Jet-Application-Version", JetVersion)

	return client.Do(req)
}

func (j *JEClient) httpPatch(url string, body any) (*http.Response, error) {
	return j.httpDO(url, body, http.MethodPatch)
}

func (j *JEClient) httpPut(url string, body any) (*http.Response, error) {
	return j.httpDO(url, body, http.MethodPut)
}

func (j *JEClient) httpPost(url string, body any) (*http.Response, error) {
	return j.httpDO(url, body, http.MethodPost)
}

// unauthorizedPost is a POST request specifically for the refresh token endpoint.
func (j *JEClient) unauthorizedPost(url string, body url.Values) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, strings.NewReader(body.Encode()))

	// The authorization is a combination of the application's UUID and name.
	basicStr := fmt.Sprintf("%s:%s", clientNames[j.Country], clientUUIDs[j.Country])
	auth := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basicStr)))

	req.Header.Set("User-Agent", UserAgent(j.DeviceModel))
	req.Header.Set("Application-Id", ApplicationID)
	req.Header.Set("Application-Version", ApplicationVersion)
	req.Header.Set("Accept-Language", languageCodes[j.Country])
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Accept-Tenant", string(j.Country))
	req.Header.Set("Accept", Accept)
	req.Header.Set("Accept-Version", AcceptVersion)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Jet-Application-Id", JetApplicationID)
	req.Header.Set("X-Jet-Application-Version", JetVersion)

	return client.Do(req)
}

func (j *JEClient) BrainTreePOST(url string, body any, headers map[string]string) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

func (j *JEClient) PayPalPOST(url string, body url.Values, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, strings.NewReader(body.Encode()))

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}
