package skip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func (c *Client) GetRestaurants() error {
	query := map[string]any{
		"operationName": "QueryRestaurantSearch",
		"variables": map[string]any{
			"city":       "toronto",
			"province":   "ON",
			"latitude":   43.7189994,
			"longitude":  -79.4753322,
			"dateTime":   0,
			"isDelivery": true,
			"search":     "c",
			"sortBy": map[string]any{
				"index": -1,
				"value": nil,
			},
			"language": "en",
			"address": map[string]any{
				"name":              "555 Rustic Rd, North York, ON M6L 1X8, Canada",
				"city":              "Toronto",
				"address":           "555 Rustic Rd, North York, ON M6L 1X8, Canada",
				"address1":          "",
				"address2":          "",
				"latitude":          43.7189994,
				"longitude":         -79.4753322,
				"verifiedAccuracy":  false,
				"useLatLongAddress": false,
				"province":          "ON",
				"postalCode":        "M6L 1X8",
			},
		},
		"extensions": map[string]any{
			"persistedQuery": map[string]any{
				"version":    1,
				"sha256Hash": "bc82691b626edbbfd55a6fde11e4cf41af93273fa0bb511c7372379941e2cc1d",
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return err
	}

	req, err := httpPost(GraphQLURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var content map[string]any
	err = json.Unmarshal(data, &content)
	if err != nil {
		return err
	}

	fmt.Println(content["data"].(map[string]any)["restaurantsList"].(map[string]any)["openRestaurants"].([]any)[0].(map[string]any)["name"])

	return nil
}
