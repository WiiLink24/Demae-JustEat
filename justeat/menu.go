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

const (
	RecommendedNameLength = 16
	NormalNameLength      = 26
)

func (j *JEClient) getCorrectMenu(menus []Menu) (*Menu, error) {
	// The menu is based on the current date and time.
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return nil, err
	}

	// Different menus can be available at different times on the same day.
	// For example, Seven Sisters KFC in London has a Lunch Menu which runs from 10 AM to 2:59 PM.
	// We have to iterate over every schedule in every menu until we find the appropriate menu.
	currentTime := time.Now().In(zone)

	// As the API gives us only the time and not absolute date, when we call time.Parse in the loop, it will return a time object of 0000-01-01 with the scheduled times.
	// Therefore, we require two time objects to compare with the scheduled times. One with the correct day of the week, and a second with the date of 0000-01-01 and current time.
	t := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond(), zone)
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
			if currentTime.Weekday().String() != schedule.DayOfWeek {
				// Not the current day, skip
				continue
			}

			for _, timeStruct := range schedule.Times {
				start, err := time.Parse("15:04:05", timeStruct.FromLocalTime)
				if err != nil {
					return nil, err
				}

				end, err := time.Parse("15:04:05", timeStruct.ToLocalTime)
				if err != nil {
					return nil, err
				}

				// We found the menu if it is the current day of the week, and the current time is not before or after the
				// start and end times.
				if !t.Before(start) && !t.After(end) {
					return &menu, nil
				}
			}
		}
	}

	return nil, nil
}

func (j *JEClient) GetRecommendedItems(id string, restaurant Restaurant) ([]demae.Item, error) {
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
		"restaurantId": restaurant.RestaurantId,
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

	resp, err = j.httpGet(fmt.Sprintf("%s/%s", j.GlobalAPIURL, restaurant.ItemsUrl))
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

	// Get possible item modifiers
	modifiers, err := j.getItemsDetails(restaurant)
	if err != nil {
		return nil, err
	}

	soldOutItems, err := j.getSoldOutItems(restaurant)
	if err != nil {
		return nil, err
	}

	menu, err := j.getCorrectMenu(restaurant.Menus)
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

		if recs["themes"] == nil {
			break
		}

		for _, rec := range recs["themes"].([]any)[0].(map[string]any)["recommendations"].([]any) {
			if rec.(map[string]any)["productId"] == item.Id {
				// Find the category for the product
				var category Category
				for _, _category := range menu.Categories {
					if slices.Contains(_category.ItemIds, item.Id) {
						category = _category
						break
					}
				}

				// Download image and process item.
				itemObj, err := j.getItem(item, id, category.Id, modifiers, items, i, soldOutItems, RecommendedNameLength)
				if err != nil {
					return nil, err
				}

				if itemObj == nil {
					continue
				}

				retItem := itemObj.Item
				retItem.XMLName = xml.Name{Local: fmt.Sprintf("container%d", i)}
				retItems = append(retItems, retItem)
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
			LinkTitle:   demae.CDATA{Value: demae.Wordwrap(category.Name, 25, 2)},
			EnabledLink: demae.CDATA{Value: 1},
			Name:        demae.CDATA{Value: category.Name},
			Info:        demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(category.Description), 50, 2)},
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

	// Get possible item modifiers
	modifiers, err := j.getItemsDetails(rest)
	if err != nil {
		return nil, err
	}

	soldOutItems, err := j.getSoldOutItems(rest)
	if err != nil {
		return nil, err
	}

	var retItems []demae.NestedItem
	i := 0
	for _, _item := range items.Items {
		if slices.Contains(category.ItemIds, _item.Id) {
			itemObj, err := j.getItem(_item, shopID, categoryID, modifiers, items, i, soldOutItems, NormalNameLength)
			if err != nil {
				return nil, err
			}

			if itemObj == nil {
				continue
			}

			retItems = append(retItems, *itemObj)
			i++
		}
	}

	return retItems, nil
}

func (j *JEClient) getItem(item Item, shopID string, categoryID string, modifiers *Modifiers, items Items, idx int, soldOutItems []string, nameWrapLen uint) (*demae.NestedItem, error) {
	// Demae has a limit of 31 characters for the item ID. As most observed IDs were UUID, a solution was to strip
	// the dashes and convert from hex to string then base64 encode. However, some IDs were observed to not be RFC
	// compliant UUIDs leading to this solution not being universal. The chosen solution I have is to generate a UUID
	// for an item and store it in Redis with the provided item ID as a key.
	var err error
	itemId := demae.CompressUUID(demae.UUID())
	if j.KeyExists(item.Id) {
		itemId, err = j.GetKey(item.Id)
		if err != nil {
			return nil, err
		}
	} else {
		err = j.SetKey(item.Id, itemId)
		if err != nil {
			return nil, err
		}
	}

	imageId := itemId
	if len(item.ImageSources) != 0 {
		j.DownloadFoodImage(item.ImageSources[0].Path, shopID, itemId)
	} else {
		imageId = "non"
	}

	if item.Description == "" {
		item.Description = "No description"
	}

	// Determine other variations of this item
	var variations []demae.ItemSize
	if item.Type == "deal" {
		// An item can be considered a "deal". We are required to select an item from the deal groups
		// before proceeding as the deal which we select impacts all modifiers.
		// We are guaranteed one "variation" as well as one "DealGroupId"
		variation := item.Variations[0]

		// We cannot support multiple DealGroups as it has a level of depth Demae does not allow.
		if len(variation.DealGroupsIds) > 1 {
			return nil, nil
		}

		for _, group := range modifiers.DealGroups {
			if slices.Contains(variation.DealGroupsIds, group.Id) {
				// We can use this deal.
				for i, itemVariation := range group.DealItemVariations {
					// We have to look up the variation in the items list.
					idx := slices.IndexFunc(items.Items, func(i Item) bool {
						return i.Id == itemVariation.DealItemVariationId
					})

					if idx == -1 {
						continue
					}

					// Demae does not give us the parent item ID even though we have to supply it.
					// We also need the Deal ID for when we add to basket.
					// Therefore, we use this format for the item code:
					// dealID|itemID|modifierID
					curItemVar := items.Items[idx]
					dealItemCode := group.Id + "|" + item.Id + "|" + curItemVar.Id
					dealItemCodeForDemae := demae.CompressUUID(demae.UUID())
					if j.KeyExists(dealItemCode) {
						dealItemCodeForDemae, err = j.GetKey(dealItemCode)
						if err != nil {
							return nil, err
						}
					} else {
						err = j.SetKey(dealItemCode, dealItemCodeForDemae)
						if err != nil {
							return nil, err
						}
					}

					variations = append(variations, demae.ItemSize{
						XMLName:   xml.Name{Local: fmt.Sprintf("item%d", i)},
						ItemCode:  demae.CDATA{Value: dealItemCodeForDemae},
						Size:      demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(curItemVar.Name), 21, 2)},
						Price:     demae.CDATA{Value: fmt.Sprintf("%.2f", variation.BasePrice+itemVariation.AdditionPrice)},
						IsSoldout: demae.CDATA{Value: demae.BoolToInt(slices.Contains(soldOutItems, curItemVar.Id) || slices.Contains(soldOutItems, item.Id))},
					})
				}
			}
		}
	} else {
		for i, variation := range item.Variations {
			name := variation.Name
			if name == "" {
				name = item.Name
			}

			itemCode := demae.CompressUUID(demae.UUID())
			if j.KeyExists(variation.Id) {
				itemCode, err = j.GetKey(variation.Id)
				if err != nil {
					return nil, err
				}
			} else {
				err = j.SetKey(variation.Id, itemCode)
				if err != nil {
					return nil, err
				}
			}

			variations = append(variations, demae.ItemSize{
				XMLName:   xml.Name{Local: fmt.Sprintf("item%d", i)},
				ItemCode:  demae.CDATA{Value: itemCode},
				Size:      demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(name), 21, 2)},
				Price:     demae.CDATA{Value: variation.BasePrice},
				IsSoldout: demae.CDATA{Value: demae.BoolToInt(slices.Contains(soldOutItems, variation.Id) || slices.Contains(soldOutItems, item.Id))},
			})
		}
	}

	if len(variations) == 0 {
		return nil, nil
	}

	return &demae.NestedItem{
		XMLName: xml.Name{Local: fmt.Sprintf("container%d", idx)},
		Name:    demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(item.Name), 26, -1)},
		Item: demae.Item{
			XMLName:    xml.Name{Local: "item"},
			MenuCode:   demae.CDATA{Value: categoryID},
			ItemCode:   demae.CDATA{Value: itemId},
			Name:       demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(item.Name), nameWrapLen, -1)},
			Price:      demae.CDATA{Value: 0},
			Info:       demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(item.Description), 36, 3)},
			Size:       nil,
			IsSelected: nil,
			Image:      demae.CDATA{Value: imageId},
			IsSoldout:  demae.CDATA{Value: demae.BoolToInt(slices.Contains(soldOutItems, item.Id))},
			SizeList: &demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "sizeList"},
				Value:   []any{variations[:]},
			},
		},
	}, nil
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

	// Determine if this is a deal product.
	justEatItemCode, err := j.GetKey(itemCode)
	if err != nil {
		return nil, 0, err
	}

	itemIDs := strings.Split(justEatItemCode, "|")
	if len(itemIDs) == 3 {
		// We need the modifier code.
		justEatItemCode = itemIDs[2]
	}

	var variation Variation
	for _, _item := range items.Items {
		// Find item
		if strings.Contains(justEatItemCode, _item.Id) {
			// Find correct variation
			idx := slices.IndexFunc(_item.Variations, func(_var Variation) bool {
				return _var.Id == justEatItemCode
			})

			variation = _item.Variations[idx]
			break
		}
	}

	err = resp.Body.Close()

	modifiers, err := j.getItemsDetails(rest)
	if err != nil {
		return nil, 0, err
	}

	var itemModifiers []demae.ItemOne
	var idx int
	for _, group := range modifiers.ModifierGroups {
		if slices.Contains(variation.ModifierGroupsIds, group.Id) {
			// Save group ID to Redis
			groupId := demae.CompressUUID(demae.UUID())
			if j.KeyExists(group.Id) {
				groupId, err = j.GetKey(group.Id)
				if err != nil {
					return nil, 0, err
				}
			} else {
				err = j.SetKey(group.Id, groupId)
				if err != nil {
					return nil, 0, err
				}
			}

			buttonType := "box"
			if group.MaxChoices == group.MinChoices {
				// In such a case we need a singular menu for each, as Demae does not allow for multiple selections
				// of mandatory items.
				buttonType = "radio"

				var modifierList []any
				for _, set := range modifiers.ModifierSets {
					if slices.Contains(group.Modifiers, set.Id) {
						// Save the item code to Redis.
						modifierId := demae.CompressUUID(demae.UUID())
						if j.KeyExists(set.Modifier.Id) {
							modifierId, err = j.GetKey(set.Modifier.Id)
							if err != nil {
								return nil, 0, err
							}
						} else {
							err = j.SetKey(set.Modifier.Id, modifierId)
							if err != nil {
								return nil, 0, err
							}
						}

						modifierList = append(modifierList, &demae.Item{
							MenuCode:  demae.CDATA{Value: groupId},
							ItemCode:  demae.CDATA{Value: modifierId},
							Name:      demae.CDATA{Value: demae.Wordwrap(set.Modifier.Name, 18, 2)},
							Price:     demae.CDATA{Value: set.Modifier.AdditionPrice},
							Info:      demae.CDATA{Value: "None yet"},
							Size:      nil,
							Image:     demae.CDATA{Value: "non"},
							IsSoldout: demae.CDATA{Value: 0},
							SizeList:  nil,
						})
					}
				}

				for _j := 0; _j < group.MaxChoices; _j++ {
					thisModifierList := make([]any, len(modifierList))
					copy(thisModifierList, modifierList)
					parent := demae.ItemOne{
						XMLName: xml.Name{Local: fmt.Sprintf("container%d", idx)},
						Info:    demae.CDATA{Value: fmt.Sprintf("Select an item.")},
						Code:    demae.CDATA{Value: groupId},
						Type:    demae.CDATA{Value: buttonType},
						Name:    demae.CDATA{Value: group.Name},
						List: demae.KVFieldWChildren{
							XMLName: xml.Name{Local: "list"},
							Value:   thisModifierList,
						},
					}

					for i, a := range parent.List.Value {
						parent.List.Value[i].(*demae.Item).ItemCode = demae.CDATA{Value: fmt.Sprintf("%s_%d", a.(*demae.Item).ItemCode.Value, _j)}
					}

					itemModifiers = append(itemModifiers, parent)
				}
			} else {
				parent := demae.ItemOne{
					XMLName: xml.Name{Local: fmt.Sprintf("container%d", idx)},
					Info:    demae.CDATA{Value: fmt.Sprintf("Max item selection is %d, minimum %d", group.MaxChoices, group.MinChoices)},
					Code:    demae.CDATA{Value: demae.CompressUUID(group.Id)},
					Type:    demae.CDATA{Value: buttonType},
					Name:    demae.CDATA{Value: group.Name},
					List: demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "list"},
					},
				}

				for _, set := range modifiers.ModifierSets {
					if slices.Contains(group.Modifiers, set.Id) {
						// Save the item code to Redis.
						modifierId := demae.CompressUUID(demae.UUID())
						if j.KeyExists(set.Modifier.Id) {
							modifierId, err = j.GetKey(set.Modifier.Id)
							if err != nil {
								return nil, 0, err
							}
						} else {
							err = j.SetKey(set.Modifier.Id, modifierId)
							if err != nil {
								return nil, 0, err
							}
						}

						parent.List.Value = append(parent.List.Value, demae.Item{
							MenuCode:  demae.CDATA{Value: groupId},
							ItemCode:  demae.CDATA{Value: modifierId},
							Name:      demae.CDATA{Value: demae.Wordwrap(set.Modifier.Name, 18, 2)},
							Price:     demae.CDATA{Value: set.Modifier.AdditionPrice},
							Info:      demae.CDATA{Value: "None yet"},
							Size:      nil,
							Image:     demae.CDATA{Value: "non"},
							IsSoldout: demae.CDATA{Value: 0},
							SizeList:  nil,
						})
					}
				}

				itemModifiers = append(itemModifiers, parent)
			}

			idx++
		}
	}

	return itemModifiers, variation.BasePrice, nil
}

func (j *JEClient) getItemsDetails(rest Restaurant) (*Modifiers, error) {
	_url := fmt.Sprintf("%s/%s", j.GlobalAPIURL, rest.ItemDetailsUrl)
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var modifiers Modifiers
	err = json.Unmarshal(body, &modifiers)
	if err != nil {
		return nil, err
	}

	return &modifiers, nil
}

func (j *JEClient) getSoldOutItems(rest Restaurant) ([]string, error) {
	// Get sold out items.
	_url := fmt.Sprintf("%s/restaurant/%s/%s/menu/dynamic", j.KongAPIURL, strings.ToLower(string(j.Country)), rest.RestaurantId)
	resp, err := j.httpGet(_url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var restaurantSummary map[string]any
	err = json.Unmarshal(body, &restaurantSummary)
	if err != nil {
		return nil, err
	}

	var offlineItems []string
	for _, a := range restaurantSummary["OfflineVariationIds"].([]any) {
		offlineItems = append(offlineItems, a.(string))
	}
	return offlineItems, nil
}
