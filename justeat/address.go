package justeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/WiiLink24/DemaeJustEat/logger"
	"golang.org/x/time/rate"
)

type NominatimResponse struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Address struct {
		Town    string `json:"town"`
		City    string `json:"city"`
		Village string `json:"string"`
	} `json:"address"`
}

// Rate limit ourselves to 1 request per second, in line with Nominatim's Usage Policy
var nominatimLimiter = rate.NewLimiter(rate.Every(time.Second), 1)

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
		if data["errors"].([]any)[0].(map[string]any)["description"].(string) == "Address cannot be geocoded" {
			// Unable to be found via Just Eat geocoder, we move onto Nominatim
			return j.getGeocodedAddressNominatim()
		}

		return 0, 0, "", errors.New(data["errors"].([]any)[0].(map[string]any)["description"].(string))
	}

	long = data["geometry"].(map[string]any)["coordinates"].([]any)[0].(float64)
	lat = data["geometry"].(map[string]any)["coordinates"].([]any)[1].(float64)
	city = data["properties"].(map[string]any)["addressLineMapping"].(map[string]any)["city"].(string)
	return
}

func (j *JEClient) getGeocodedAddressNominatim() (long float64, lat float64, city string, e error) {
	var err error

	street := url.QueryEscape(j.Address)
	postalCode := url.QueryEscape(j.PostalCode)

	var bodyString string
	// Check if address is cached
	if j.KeyExists(street + ", " + postalCode) {
		bodyString, err = j.GetKey(street + ", " + postalCode)
		if err != nil {
			return 0, 0, "", err
		}
	} else {
		err := nominatimLimiter.Wait(j.Context)
		if err != nil {
			return 0, 0, "", err
		}

		url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?street=%s&postalcode=%s&format=json&limit=1&addressdetails=1", street, postalCode)
		body, err := HttpGet(url)
		if err != nil {
			return 0, 0, "", err
		}
		bodyString = string(body)

		err = j.SetKey(street+", "+postalCode, bodyString)
		if err != nil {
			return 0, 0, "", err
		}
	}

	var results []NominatimResponse
	err = json.Unmarshal([]byte(bodyString), &results)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, "", fmt.Errorf("failed to find address - make sure it is set correctly")
	}

	long, err = strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse longitude: %w", err)
	}
	lat, err = strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse latitude: %w", err)
	}

	if len(results[0].Address.Village) > 0 {
		city = results[0].Address.Village
	} else if len(results[0].Address.Town) > 0 {
		city = results[0].Address.Town
	} else {
		city = results[0].Address.City
	}

	return
}
