package skip

import (
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"io"
)

func (c *Client) getGeocodedAddress() (long float64, lat float64, city string, e error) {
	payload := map[string]any{
		"addressLines": []string{
			c.Address,
			"",
			"",
			c.PostalCode,
		},
	}

	// We can use the Just Eat geocoder.
	_url := fmt.Sprintf("%s/geocode/%s", justeat.KongAPIURLs[justeat.UnitedKingdom], justeat.UnitedKingdom)
	resp, err := httpPost(_url, payload)
	if err != nil {
		return 0, 0, "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// Retrieve longitude and latitude
	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, 0, "", err
	}

	long = data["geometry"].(map[string]any)["coordinates"].([]any)[0].(float64)
	lat = data["geometry"].(map[string]any)["coordinates"].([]any)[1].(float64)
	city = data["properties"].(map[string]any)["addressLineMapping"].(map[string]any)["city"].(string)
	return
}
