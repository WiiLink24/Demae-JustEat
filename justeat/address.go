package justeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/WiiLink24/DemaeJustEat/logger"
)

func (j *JEClient) getGeocodedAddress() (long float64, lat float64, city string, e error) {
	payload := map[string]any{
		"addressLines": []string{
			j.Address,
			"",
			"",
			j.PostalCode,
		},
	}

	_url := fmt.Sprintf("%s/geocode/%s", j.KongAPIURL, j.Country)
	resp, err := j.httpPost(_url, payload)
	if err != nil {
		return 0, 0, "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error(Address, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", err
	}

	// Retrieve longitude and latitude
	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, 0, "", err
	}

	if data["errors"] != nil {
		return 0, 0, "", errors.New(data["errors"].([]any)[0].(map[string]any)["description"].(string))
	}

	long = data["geometry"].(map[string]any)["coordinates"].([]any)[0].(float64)
	lat = data["geometry"].(map[string]any)["coordinates"].([]any)[1].(float64)
	city = data["properties"].(map[string]any)["addressLineMapping"].(map[string]any)["city"].(string)
	return
}
