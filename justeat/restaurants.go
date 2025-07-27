package justeat

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	dealsMap  = map[string]string{}
	dealsLock = &sync.Mutex{}
)

func (j *JEClient) GetBareRestaurants() (c []demae.CategoryCode, e error) {
	long, lat, _, err := j.getGeocodedAddress()
	if err != nil {
		e = err
		return
	}

	queryParams := url.Values{
		"latitude":              {strconv.FormatFloat(lat, 'f', -1, 64)},
		"longitude":             {strconv.FormatFloat(long, 'f', -1, 64)},
		"ratingsOutOfFive":      {"true"},
		"include-test-partners": {"false"},
		"serviceType":           {"delivery"},
	}

	_url := fmt.Sprintf("%s/discovery/%s/restaurants/enriched?%s", j.KongAPIURL, j.Country, queryParams.Encode())
	resp, err := j.httpGet(_url)
	if err != nil {
		e = err
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e = err
		return
	}

	// Decode to map and extract
	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		e = err
		return
	}

	for _, restaurant := range data["restaurants"].([]any) {
		// Parse the cuisine types
		for _, cuisine := range restaurant.(map[string]any)["cuisines"].([]any) {
			for code, strings := range categoryTypes {
				if slices.Contains(strings, cuisine.(map[string]any)["uniqueName"].(string)) && !slices.Contains(c, code) && restaurant.(map[string]any)["isDelivery"].(bool) {
					c = append(c, code)
				}
			}
		}
	}

	return
}

func (j *JEClient) GetRestaurants(code demae.CategoryCode) ([]demae.BasicShop, error) {
	long, lat, _, err := j.getGeocodedAddress()
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{
		"latitude":              {strconv.FormatFloat(lat, 'f', -1, 64)},
		"longitude":             {strconv.FormatFloat(long, 'f', -1, 64)},
		"ratingsOutOfFive":      {"true"},
		"include-test-partners": {"false"},
		"serviceType":           {"delivery"},
	}

	_url := fmt.Sprintf("%s/discovery/%s/restaurants/enriched?%s", j.KongAPIURL, j.Country, queryParams.Encode())
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// Decode to map and extract
	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	var restaurants []demae.BasicShop
	var numOfRestaurants int
	for _, restaurant := range data["restaurants"].([]any) {
		if numOfRestaurants == MaxNumberOfRestaurants {
			break
		}

		// Ensure this restaurant is the requested category type
		var isCategory bool
		for _, cuisine := range restaurant.(map[string]any)["cuisines"].([]any) {
			if isCategory {
				break
			}

			for _code, s := range categoryTypes {
				if _code == code {
					if slices.Contains(s, cuisine.(map[string]any)["uniqueName"].(string)) {
						// If it does contain we are done here
						isCategory = true
						break
					}
				}
			}
		}

		if !isCategory {
			continue
		}

		if !restaurant.(map[string]any)["isDelivery"].(bool) {
			break
		}

		// Store minimum delivery price in the cache.
		shopCode := restaurant.(map[string]any)["uniqueName"].(string)
		minDeliveryPrice := restaurant.(map[string]any)["minimumDeliveryValue"].(float64)

		// Download image
		imgUrl := restaurant.(map[string]any)["logoUrl"].(string)
		j.DownloadLogo(imgUrl, restaurant.(map[string]any)["uniqueName"].(string))

		// Get any available offers
		dealsStr := ""
		for _, deals := range restaurant.(map[string]any)["deals"].([]any) {
			if deals.(map[string]any)["description"].(string) == "" {
				continue
			}

			dealsStr += deals.(map[string]any)["description"].(string) + "\n"
		}

		// Remove last newline and put in map.
		dealsStr = strings.TrimRight(dealsStr, "\n")
		dealsLock.Lock()
		dealsMap[shopCode] = dealsStr
		dealsLock.Unlock()

		restaurants = append(restaurants, demae.BasicShop{
			ShopCode:    demae.CDATA{Value: shopCode},
			HomeCode:    demae.CDATA{Value: shopCode},
			Name:        demae.CDATA{Value: demae.Wordwrap(restaurant.(map[string]any)["name"].(string), 24, -1)},
			Catchphrase: demae.CDATA{Value: "None"},
			MinPrice:    demae.CDATA{Value: minDeliveryPrice},
			Yoyaku:      demae.CDATA{Value: 1},
			Activate:    demae.CDATA{Value: "on"},
			WaitTime:    demae.CDATA{Value: restaurant.(map[string]any)["availability"].(map[string]any)["delivery"].(map[string]any)["etaMinutes"].(map[string]any)["rangeLower"]},
			PaymentList: demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "paymentList"},
				Value: []any{
					demae.KVField{
						XMLName: xml.Name{Local: "athing"},
						Value:   "Fox Card",
					},
				},
			},
			ShopStatus: demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "shopStatus"},
				Value: []any{
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "status"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   demae.BoolToInt(restaurant.(map[string]any)["availability"].(map[string]any)["delivery"].(map[string]any)["isOpen"].(bool)),
							},
						},
					},
				},
			},
		})

		numOfRestaurants++
	}

	return restaurants, nil
}

func (j *JEClient) GetRestaurant(id string) (*demae.ShopOne, error) {
	_url := fmt.Sprintf("%s/%s_%s_manifest.json", j.GlobalAPIURL, id, strings.ToLower(string(j.Country)))
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// Decode to map and extract
	var rest Restaurant
	err = json.Unmarshal(body, &rest)
	if err != nil {
		return nil, err
	}

	activate := "on"
	if rest.RestaurantInfo.IsOffline {
		activate = "off"
	}

	menu, err := j.getCorrectMenu(rest.Menus)
	if err != nil {
		return nil, err
	}

	// With the correct menu, we can pull the times that the restaurant is open.
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return nil, err
	}

	t := time.Now().In(zone)
	var times []demae.KVFieldWChildren
	var waitingTime int
	for _, schedule := range menu.Schedules {
		if schedule.DayOfWeek != t.Weekday().String() {
			continue
		}

		// Correct day, get operating times.
		waitingTime = int(schedule.OrderTimeSchedule[0].LowerBound)

		for i, timeStruct := range schedule.Times {
			times = append(times, demae.KVFieldWChildren{
				XMLName: xml.Name{Local: fmt.Sprintf("values%d", i)},
				Value: []any{
					demae.KVField{
						XMLName: xml.Name{Local: "start"},
						Value:   timeStruct.FromLocalTime,
					},
					demae.KVField{
						XMLName: xml.Name{Local: "end"},
						Value:   timeStruct.ToLocalTime,
					},
					demae.KVField{
						XMLName: xml.Name{Local: "holiday"},
						Value:   "n",
					},
				},
			})
		}
	}

	var orderTimes []demae.KVFieldWChildren
	basketId := j.FakeBasket(id, menu.MenuGroupId)
	if basketId != "" {
		orderTimes, err = j.GetAvailableTimes(basketId)
		if err != nil {
			return nil, err
		}

		orderTimes = nil
	}

	recommendations, err := j.GetRecommendedItems(id, rest)
	if err != nil {
		return nil, err
	}

	dealsLock.Lock()
	deals := dealsMap[id]
	dealsLock.Unlock()

	return &demae.ShopOne{
		CategoryCode:  demae.CDATA{Value: "01"},
		Address:       demae.CDATA{Value: rest.RestaurantInfo.Location.Address},
		Information:   demae.CDATA{Value: demae.Wordwrap(rest.RestaurantInfo.Description, 24, -1)},
		Attention:     demae.CDATA{Value: "None"},
		Amenity:       demae.CDATA{Value: deals},
		MenuListCode:  demae.CDATA{Value: 1},
		Activate:      demae.CDATA{Value: activate},
		WaitTime:      demae.CDATA{Value: waitingTime},
		TimeOrder:     demae.CDATA{Value: "y"},
		Tel:           demae.CDATA{Value: "None"},
		YoyakuMinDate: demae.CDATA{Value: 1},
		YoyakuMaxDate: demae.CDATA{Value: 30},
		PaymentList: demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "paymentList"},
			Value: []any{
				demae.KVField{
					XMLName: xml.Name{Local: "athing"},
					Value:   "Fox Card",
				},
			},
		},
		ShopStatus: demae.ShopStatus{
			Hours: demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "hours"},
				Value: []any{
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "all"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "message"},
								Value:   "hi",
							},
						},
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "today"},
						Value: []any{
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "values"},
								Value: []any{
									times[:],
								},
							},
						},
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "delivery"},
						Value: []any{
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "values"},
								Value: []any{
									times[:],
								},
							},
						},
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "selList"},
						Value: []any{
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "values"},
								Value: []any{
									orderTimes[:],
								},
							},
						},
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "status"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   demae.BoolToInt(true),
							},
						},
					},
				},
			},
			Interval: demae.CDATA{Value: 5},
			Holiday:  demae.CDATA{Value: "No information."},
		},
		RecommendedItemList: demae.KVFieldWChildren{
			Value: []any{
				recommendations[:],
				demae.Item{
					XMLName:   xml.Name{Local: "container4"},
					MenuCode:  demae.CDATA{Value: 10},
					ItemCode:  demae.CDATA{Value: 1},
					Name:      demae.CDATA{Value: "Pizza"},
					Price:     demae.CDATA{Value: 10},
					Info:      demae.CDATA{Value: "Fresh"},
					Size:      &demae.CDATA{Value: 1},
					Image:     demae.CDATA{Value: "PIZZA"},
					IsSoldout: demae.CDATA{Value: 0},
					SizeList: &demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "sizeList"},
						Value: []any{
							demae.ItemSize{
								XMLName:   xml.Name{Local: "item1"},
								ItemCode:  demae.CDATA{Value: 1},
								Size:      demae.CDATA{Value: 1},
								Price:     demae.CDATA{Value: 10},
								IsSoldout: demae.CDATA{Value: 0},
							},
						},
					},
				},
			},
		},
	}, nil
}
