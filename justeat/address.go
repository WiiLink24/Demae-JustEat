package justeat

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

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
	err := nominatimLimiter.Wait(j.Context)
	if err != nil {
		return 0, 0, "", err
	}

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
		return 0, 0, "", fmt.Errorf("no location found")
	}

	long, err = strconv.ParseFloat(results[0].Lon, 64)
	lat, err = strconv.ParseFloat(results[0].Lat, 64)
	if len(results[0].Address.Village) > 0 {
		city = results[0].Address.Village
	} else if len(results[0].Address.Town) > 0 {
		city = results[0].Address.Town
	} else {
		city = results[0].Address.City
	}

	return
}
