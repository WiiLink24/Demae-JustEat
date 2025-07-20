package justeat

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"io"
	"slices"
	"strings"
	"time"
)

func (j *JEClient) getCorrectMenu(menus []Menu) (Menu, error) {
	// The menu is based on the current date and time.
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return Menu{}, err
	}

	// Different menus can be available at different times on the same day.
	// For example, Seven Sisters KFC in London has a Lunch Menu which runs from 10 AM to 2:59 PM.
	// We have to iterate over every schedule in every menu until we find the appropriate menu.
	t := time.Now().In(zone)

	// As the API gives us only the time and not absolute date, we have to default to Go's default time of 0000-01-01
	t = time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), zone)
	for _, menu := range menus {
		// Skip non delivery menus
		hasDelivery := false
		for _, serviceType := range menu.ServiceTypes {
			if serviceType == "delivery" {
				hasDelivery = true
			}
		}

		if !hasDelivery {
			continue
		}

		for _, schedule := range menu.Schedules {
			if t.Weekday().String() != schedule.DayOfWeek {
				// Not the current day, skip
				continue
			}

			// TODO: Index 0 might be bad lmao
			start, err := time.Parse("15:04:05", schedule.Times[0].FromLocalTime)
			if err != nil {
				return Menu{}, err
			}

			end, err := time.Parse("15:04:05", schedule.Times[0].ToLocalTime)
			if err != nil {
				return Menu{}, err
			}

			// We found the menu if it is the current day of the week, and the current time is not before or after the
			// start and end times.
			if !t.Before(start) && !t.After(end) {
				return menu, nil
			}
		}
	}

	return Menu{}, nil
}

func (j *JEClient) GetRecommendedItems(restaurantId string, itemsUrl string, longRestaurantId string) ([]demae.Item, error) {
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return nil, err
	}

	// Get current day of the week
	payload := map[string]any{
		"orderRequestDetails": map[string]any{
			"dayOfWeek":      int(time.Now().In(zone).Weekday()),
			"orderedForTime": time.Now().In(zone).Format("15:04:05"),
			"serviceType":    "delivery",
		},
		"restaurantId": restaurantId,
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/recommendations/%s/dishes/menu", j.KongAPIURL, strings.ToLower(string(j.Country))), payload)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var recs map[string]any
	err = json.Unmarshal(body, &recs)
	if err != nil {
		return nil, err
	}

	resp, err = j.httpGet(fmt.Sprintf("%s/%s", j.GlobalAPIURL, itemsUrl))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items Items
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, err
	}

	var retItems []demae.Item
	i := 0
	for _, item := range items.Items {
		// Demae only allows 3 recommended items.
		if i == 3 {
			break
		}

		for _, rec := range recs["themes"].([]any)[0].(map[string]any)["recommendations"].([]any) {
			if rec.(map[string]any)["productId"] == item.Id {
				// Do the thing
				j.DownloadFoodImage(item.ImageSources[0].Path, longRestaurantId, item.Id)

				if item.Description == "" {
					item.Description = "No description"
				}

				// Determine other variations of this item
				var variations []demae.ItemSize
				for i, variation := range item.Variations {
					name := variation.Name
					if name == "" {
						name = item.Name
					}

					variations = append(variations, demae.ItemSize{
						XMLName:   xml.Name{Local: fmt.Sprintf("item%d", i)},
						ItemCode:  demae.CDATA{Value: variation.Id},
						Size:      demae.CDATA{Value: demae.RemoveInvalidCharacters(name)},
						Price:     demae.CDATA{Value: variation.BasePrice},
						IsSoldout: demae.CDATA{Value: 0},
					})
				}

				retItems = append(retItems, demae.Item{
					XMLName:    xml.Name{Local: fmt.Sprintf("container%d", i)},
					MenuCode:   demae.CDATA{Value: 0},
					ItemCode:   demae.CDATA{Value: item.Id},
					Name:       demae.CDATA{Value: demae.RemoveInvalidCharacters(item.Name)},
					Price:      demae.CDATA{Value: 0},
					Info:       demae.CDATA{Value: demae.RemoveInvalidCharacters(item.Description)},
					Size:       nil,
					IsSelected: nil,
					Image:      demae.CDATA{Value: item.Id},
					IsSoldout:  demae.CDATA{Value: 0},
					SizeList: &demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "sizeList"},
						Value:   []any{variations[:]},
					},
				})
				i++
				break
			}
		}
	}

	return retItems, nil
}

func (j *JEClient) GetMenuCategories(id string) ([]demae.Menu, error) {
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

	menu, err := j.getCorrectMenu(rest.Menus)
	if err != nil {
		return nil, err
	}

	var menus []demae.Menu
	for i, category := range menu.Categories {
		if category.Description == "" {
			category.Description = "No description"
		}

		menus = append(menus, demae.Menu{
			XMLName:     xml.Name{Local: fmt.Sprintf("container_%d", i)},
			MenuCode:    demae.CDATA{Value: category.Id},
			LinkTitle:   demae.CDATA{Value: category.Name},
			EnabledLink: demae.CDATA{Value: 1},
			Name:        demae.CDATA{Value: category.Name},
			Info:        demae.CDATA{Value: ""}, // category.Description
			SetNum:      demae.CDATA{Value: 1},
			LunchMenuList: struct {
				IsLunchTimeMenu demae.CDATA `xml:"isLunchTimeMenu"`
				Hour            demae.KVFieldWChildren
				IsOpen          demae.CDATA `xml:"isOpen"`
				Message         demae.CDATA `xml:"message"`
			}{
				IsLunchTimeMenu: demae.CDATA{Value: demae.BoolToInt(false)},
				Hour: demae.KVFieldWChildren{
					XMLName: xml.Name{Local: "hour"},
					Value: []any{
						demae.KVField{
							XMLName: xml.Name{Local: "start"},
							Value:   "00:00:00",
						},
						demae.KVField{
							XMLName: xml.Name{Local: "end"},
							Value:   "24:59:59",
						},
					},
				},
				IsOpen:  demae.CDATA{Value: demae.BoolToInt(true)},
				Message: demae.CDATA{Value: "Where does this show up?"},
			},
		})
	}

	return menus, nil
}

func (j *JEClient) GetMenuItems(shopID, categoryID string) ([]demae.NestedItem, error) {
	_url := fmt.Sprintf("%s/%s_%s_manifest.json", j.GlobalAPIURL, shopID, strings.ToLower(string(j.Country)))
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rest Restaurant
	err = json.Unmarshal(body, &rest)
	if err != nil {
		return nil, err
	}

	menu, err := j.getCorrectMenu(rest.Menus)
	if err != nil {
		return nil, err
	}

	// Find our requested category
	var category Category
	for _, _category := range menu.Categories {
		if _category.Id == categoryID {
			category = _category
		}
	}

	// Close previous response body
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	_url = fmt.Sprintf("%s/%s", j.GlobalAPIURL, rest.ItemsUrl)
	resp, err = j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items Items
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, err
	}

	var retItems []demae.NestedItem
	a := 0
	for _, _item := range items.Items {
		if slices.Contains(category.ItemIds, _item.Id) {
			if len(_item.ImageSources) != 0 {
				j.DownloadFoodImage(_item.ImageSources[0].Path, shopID, _item.Id)
			}

			if _item.Description == "" {
				_item.Description = "No description"
			}

			// Determine other variations of this item
			var variations []demae.ItemSize
			for i, variation := range _item.Variations {
				name := variation.Name
				if name == "" {
					name = _item.Name
				}

				variations = append(variations, demae.ItemSize{
					XMLName:   xml.Name{Local: fmt.Sprintf("item%d", i)},
					ItemCode:  demae.CDATA{Value: variation.Id},
					Size:      demae.CDATA{Value: demae.RemoveInvalidCharacters(name)},
					Price:     demae.CDATA{Value: variation.BasePrice},
					IsSoldout: demae.CDATA{Value: 0},
				})
			}

			retItems = append(retItems, demae.NestedItem{
				XMLName: xml.Name{Local: fmt.Sprintf("container%d", a)},
				Name:    demae.CDATA{Value: demae.RemoveInvalidCharacters(_item.Name)},
				Item: demae.Item{
					XMLName:    xml.Name{Local: "item"},
					MenuCode:   demae.CDATA{Value: categoryID},
					ItemCode:   demae.CDATA{Value: _item.Id},
					Name:       demae.CDATA{Value: demae.RemoveInvalidCharacters(_item.Name)},
					Price:      demae.CDATA{Value: 0},
					Info:       demae.CDATA{Value: ""}, // demae.RemoveInvalidCharacters(_item.Description)
					Size:       nil,
					IsSelected: nil,
					Image:      demae.CDATA{Value: _item.Id},
					IsSoldout:  demae.CDATA{Value: 0},
					SizeList: &demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "sizeList"},
						Value:   []any{variations[:]},
					},
				},
			})
			a++
		}
	}

	return retItems, nil
}

func (j *JEClient) GetItemData(shopID, categoryID, itemCode string) ([]demae.ItemOne, float64, error) {
	_url := fmt.Sprintf("%s/%s_%s_manifest.json", j.GlobalAPIURL, shopID, strings.ToLower(string(j.Country)))
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var rest Restaurant
	err = json.Unmarshal(body, &rest)
	if err != nil {
		return nil, 0, err
	}

	err = resp.Body.Close()

	_url = fmt.Sprintf("%s/%s", j.GlobalAPIURL, rest.ItemsUrl)
	resp, err = j.httpGet(_url)
	if err != nil {
		return nil, 0, err
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var items Items
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, 0, err
	}

	var variation Variation
	for _, _item := range items.Items {
		// Find item
		if strings.Contains(itemCode, _item.Id) {
			// Find correct variation
			idx := slices.IndexFunc(_item.Variations, func(_var Variation) bool {
				return _var.Id == itemCode
			})

			variation = _item.Variations[idx]
			break
		}
	}

	err = resp.Body.Close()

	_url = fmt.Sprintf("%s/%s", j.GlobalAPIURL, rest.ItemDetailsUrl)
	resp, err = j.httpGet(_url)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var modifiers Modifiers
	err = json.Unmarshal(body, &modifiers)
	if err != nil {
		return nil, 0, err
	}

	var itemModifiers []demae.ItemOne
	for i, group := range modifiers.ModifierGroups {
		if slices.Contains(variation.ModifierGroupsIds, group.Id) {
			buttonType := "box"
			if group.MaxChoices == group.MinChoices {
				buttonType = "radio"
			}

			parent := demae.ItemOne{
				XMLName: xml.Name{Local: fmt.Sprintf("container%d", i)},
				Info:    demae.CDATA{Value: fmt.Sprintf("Max item selection is %d, minimum %d", group.MaxChoices, group.MinChoices)},
				Code:    demae.CDATA{Value: group.Id},
				Type:    demae.CDATA{Value: buttonType},
				Name:    demae.CDATA{Value: group.Name},
				List: demae.KVFieldWChildren{
					XMLName: xml.Name{Local: "list"},
				},
			}

			for _, set := range modifiers.ModifierSets {
				if slices.Contains(group.Modifiers, set.Id) {
					parent.List.Value = append(parent.List.Value, demae.Item{
						MenuCode:  demae.CDATA{Value: set.Modifier.Id},
						ItemCode:  demae.CDATA{Value: set.Modifier.Id},
						Name:      demae.CDATA{Value: set.Modifier.Name},
						Price:     demae.CDATA{Value: set.Modifier.AdditionPrice},
						Info:      demae.CDATA{Value: "None yet"},
						Size:      nil,
						Image:     demae.CDATA{Value: "none"},
						IsSoldout: demae.CDATA{Value: 0},
						SizeList:  nil,
					})
				}
			}

			itemModifiers = append(itemModifiers, parent)
		}
	}

	return itemModifiers, variation.BasePrice, nil
}
