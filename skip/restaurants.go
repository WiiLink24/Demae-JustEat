package skip

import (
	"encoding/json"
	"fmt"
	"io"
)

func (c *Client) GetRestaurants() error {
	query := map[string]any{
		"operationName": "QueryRestaurantsCuisinesList",
		"variables": map[string]any{
			"city":          "toronto",
			"province":      "ON",
			"latitude":      43.7188732,
			"longitude":     -79.4753502,
			"isDelivery":    true,
			"dateTime":      0,
			"search":        "",
			"language":      "en",
			"orderType":     "DELIVERY",
			"withNewImages": true,
		},
		"extensions": map[string]any{
			"persistedQuery": map[string]any{
				"version":    1,
				"sha256Hash": "64ff8a1e704dceb0066b99ff7ce9d9e0b101ad70d853f097c37ba21314188069",
			},
		},
	}

	req, err := httpPost(GraphQLURL, query)
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

	fmt.Println(content)

	return nil
}
