package justeat

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (j *JEClient) GetMenuGroupID(shopID string) (string, error) {
	_url := fmt.Sprintf("%s/%s_%s_manifest.json", j.GlobalAPIURL, shopID, strings.ToLower(string(j.Country)))
	resp, err := j.httpGet(_url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// Decode to map and extract
	var rest Restaurant
	err = json.Unmarshal(body, &rest)
	if err != nil {
		return "", err
	}

	menu, err := j.getCorrectMenu(rest.Menus)
	if err != nil {
		return "", err
	}

	return menu.MenuGroupId, nil
}

func formProduct(r *http.Request) (*Product, error) {
	itemCode := r.PostForm.Get("itemCode")
	quantityStr := r.PostForm.Get("quantity")

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		return nil, err
	}

	var modifierGroups []ModifierGroup
	for items, _ := range r.PostForm {
		if strings.Contains(items, "option") {
			// Extract the topping type and code
			var groupID string
			var modifierID string
			for i, s := range strings.Split(items, "[") {
				switch i {
				case 0:
					continue
				case 1:
					// Modifier ID
					modifierID = strings.Split(s, "]")[0]
				case 2:
					// Modifier Group ID
					groupID = strings.Split(s, "]")[0]
				}
			}

			item := Modifier{
				ID:       modifierID,
				Quantity: 1,
			}

			// Find group if it exists
			idx := slices.IndexFunc(modifierGroups, func(group ModifierGroup) bool {
				return group.GroupId == groupID
			})

			if idx == -1 {
				// Not found, create.
				modifierGroups = append(modifierGroups, ModifierGroup{
					GroupId:   groupID,
					Modifiers: []Modifier{item},
				})
			} else {
				modifierGroups[idx].Modifiers = append(modifierGroups[idx].Modifiers, item)
			}
		}
	}

	product := Product{
		Date:               time.Now().UTC().Format("2006-01-02T15:01:05.000Z"),
		ProductId:          itemCode,
		Quantity:           quantity,
		ModifierGroups:     modifierGroups,
		RemovedIngredients: nil,
	}

	return &product, nil
}

func (j *JEClient) CreateBasket(r *http.Request) (string, error) {
	shopCode := r.PostForm.Get("shopCode")
	c, err := j.GetMenuGroupID(shopCode)
	if err != nil {
		return "", err
	}

	long, lat, _, err := j.getGeocodedAddress()
	if err != nil {
		return "", err
	}

	product, err := formProduct(r)
	if err != nil {
		return "", err
	}

	basket := Basket{
		RestaurantSEOName: shopCode,
		MenuGroupId:       c,
		ServiceType:       "delivery",
		Products:          []Product{*product},
		OrderDetails: OrderDetails{
			Location: Location{
				ZipCode: j.PostalCode,
				GeoLocation: GeoLocation{
					Latitude:  lat,
					Longitude: long,
				},
			},
		},
		Consents:   []any{},
		Deals:      []any{},
		BasketMode: "None",
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/basket", j.KongAPIURL), basket)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var b struct {
		BasketId string `json:"BasketId"`
	}

	err = json.Unmarshal(body, &b)
	return b.BasketId, err
}

func (j *JEClient) FakeBasket(shopCode, menuGroupId string) string {
	long, lat, _, err := j.getGeocodedAddress()
	if err != nil {
		return ""
	}

	basket := Basket{
		RestaurantSEOName: shopCode,
		MenuGroupId:       menuGroupId,
		ServiceType:       "delivery",
		Products:          []Product{},
		OrderDetails: OrderDetails{
			Location: Location{
				ZipCode: j.PostalCode,
				GeoLocation: GeoLocation{
					Latitude:  lat,
					Longitude: long,
				},
			},
		},
		Consents:   []any{},
		Deals:      []any{},
		BasketMode: "None",
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/basket", j.KongAPIURL), basket)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var b struct {
		BasketId string `json:"BasketId"`
	}

	err = json.Unmarshal(body, &b)
	return b.BasketId
}

func (j *JEClient) EditBasket(basketId string, r *http.Request) error {
	product, err := formProduct(r)
	if err != nil {
		return err
	}

	edit := BasketEdit{
		BasketId: basketId,
		Product: BasketStatus{
			Added:   []Product{*product},
			Updated: nil,
			Removed: nil,
		},
		Deal: BasketStatus{},
	}

	_, err = j.httpPut(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId), edit)
	return err
}

func (j *JEClient) RemoveItem(basketId string, productId string, r *http.Request) error {
	remove := BasketEdit{
		BasketId: basketId,
		Product: BasketStatus{
			Added:   nil,
			Updated: nil,
			Removed: []BasketRemoval{
				{Date: time.Now().UTC().Format("2006-01-02T15:01:05.000Z"), BasketProductId: productId},
			},
		},
		Deal: BasketStatus{},
	}

	resp, err := j.httpPut(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId), remove)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return err
}

// getBasket returns the basket object from Just Eat.
func (j *JEClient) getBasket(basketId string) (BasketData, error) {
	resp, err := j.httpGet(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId))
	if err != nil {
		return BasketData{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var summary BasketData
	err = json.Unmarshal(body, &summary)
	return summary, err
}

// GetBasket returns the basket in a Demae usable structure.
func (j *JEClient) GetBasket(basketId string, r *http.Request) ([]any, error) {
	summary, err := j.getBasket(basketId)
	if err != nil {
		return nil, err
	}

	var basketItems []demae.BasketItem
	for i, product := range summary.BasketSummary.Products {
		// First group the modifiers
		var modifiers []any
		for i, option := range product.ModifierGroups {
			group := demae.ItemOne{
				XMLName: xml.Name{Local: fmt.Sprintf("container%d", i)},
				Info:    demae.CDATA{Value: ""},
				Code:    demae.CDATA{Value: option.GroupId},
				Type:    demae.CDATA{Value: 0},
				Name:    demae.CDATA{Value: fmt.Sprintf("Modifier %d", i+1)},
				List:    demae.KVFieldWChildren{},
			}

			for _, modifier := range option.Modifiers {
				group.List.Value = append(group.List.Value, demae.Item{
					MenuCode:   demae.CDATA{Value: modifier.ID},
					ItemCode:   demae.CDATA{Value: modifier.ID},
					Name:       demae.CDATA{Value: modifier.Name},
					Price:      demae.CDATA{Value: 0},
					Info:       demae.CDATA{Value: 0},
					IsSelected: &demae.CDATA{Value: demae.BoolToInt(true)},
					Image:      demae.CDATA{Value: 0},
					IsSoldout:  demae.CDATA{Value: demae.BoolToInt(false)},
				})
			}

			modifiers = append(modifiers, group)
		}

		priceStr := fmt.Sprintf("$%.2f", product.UnitPrice)
		amountStr := fmt.Sprintf("$%.2f", product.TotalPrice)
		basketItems = append(basketItems, demae.BasketItem{
			XMLName:       xml.Name{Local: fmt.Sprintf("container%d", i)},
			BasketNo:      demae.CDATA{Value: product.ProductId},
			MenuCode:      demae.CDATA{Value: 1},
			ItemCode:      demae.CDATA{Value: product.ProductId},
			Name:          demae.CDATA{Value: demae.RemoveInvalidCharacters(product.Name)},
			Price:         demae.CDATA{Value: priceStr},
			Size:          demae.CDATA{Value: ""},
			IsSoldout:     demae.CDATA{Value: demae.BoolToInt(false)},
			Quantity:      demae.CDATA{Value: product.Quantity},
			SubTotalPrice: demae.CDATA{Value: amountStr},
			Menu: demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "Menu"},
				Value: []any{
					demae.KVField{
						XMLName: xml.Name{Local: "name"},
						Value:   "Menu",
					},
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "lunchMenuList"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "isLunchTimeMenu"},
								Value:   demae.BoolToInt(false),
							},
							demae.KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   demae.BoolToInt(true),
							},
						},
					},
				},
			},
			OptionList: demae.KVFieldWChildren{
				XMLName: xml.Name{Local: ""},
				Value:   modifiers,
			},
		})
	}

	basketPrice := demae.KVField{
		XMLName: xml.Name{Local: "basketPrice"},
		Value:   summary.BasketSummary.BasketTotals.Subtotal,
	}

	chargePrice := demae.KVField{
		XMLName: xml.Name{Local: "chargePrice"},
		Value:   summary.BasketSummary.BasketTotals.Subtotal,
	}

	totalPrice := demae.KVField{
		XMLName: xml.Name{Local: "totalPrice"},
		Value:   summary.BasketSummary.BasketTotals.Total,
	}

	cart := demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "List"},
		Value:   []any{basketItems[:]},
	}

	status := demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "Status"},
		Value: []any{
			demae.KVField{
				XMLName: xml.Name{Local: "isOrder"},
				Value:   demae.BoolToInt(true),
			},
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "messages"},
				Value: []any{demae.KVField{
					XMLName: xml.Name{Local: "hey"},
					Value:   "how are you?",
				}},
			},
		},
	}

	return []any{
		basketPrice,
		chargePrice,
		totalPrice,
		status,
		cart,
	}, nil
}
