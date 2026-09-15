package justeat

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/logger"
)

func (j *JEClient) GetMenuGroupID(shopID string) (string, error) {
	_url := fmt.Sprintf("%s/%s_%s_manifest.json", j.GlobalAPIURL, shopID, strings.ToLower(string(j.Country)))
	resp, err := j.httpGet(_url)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

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

func (j *JEClient) formProduct(r *http.Request, itemCode string, quantity int) (*Product, error) {
	type ModifierPreAdd struct {
		GroupId    string
		ModifierId string
		Quantity   int
	}

	modifierMap := make(map[string]ModifierPreAdd)
	var modifierGroups []ModifierGroup
	var err error
	for items := range r.PostForm {
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
					groupID, err = j.GetKey(strings.Split(s, "]")[0])
					if err != nil {
						return nil, err
					}
				case 2:
					// Modifier Group ID
					// Can be possible to have duplicates of a modifier. In this case we want to increment the quantity.
					isDupe := false
					notModifiedId := strings.Split(s, "]")[0]
					if notModifiedId[len(notModifiedId)-2] == '_' {
						// We got one. We should also strip the suffix.
						notModifiedId = notModifiedId[:len(notModifiedId)-2]
						isDupe = true
					}

					modifierID, err = j.GetKey(notModifiedId)
					if err != nil {
						return nil, err
					}

					if isDupe {
						if m, ok := modifierMap[modifierID]; ok {
							m.Quantity++
						} else {
							modifierMap[modifierID] = ModifierPreAdd{
								GroupId:    groupID,
								ModifierId: modifierID,
								Quantity:   1,
							}
						}
					}
				}
			}
		}
	}

	for _, add := range modifierMap {
		item := Modifier{
			ID:       add.ModifierId,
			Quantity: add.Quantity,
		}

		// Find group if it exists
		idx := slices.IndexFunc(modifierGroups, func(group ModifierGroup) bool {
			return group.GroupId == add.GroupId
		})

		if idx == -1 {
			// Not found, create.
			modifierGroups = append(modifierGroups, ModifierGroup{
				GroupId:   add.GroupId,
				Modifiers: []Modifier{item},
			})
		} else {
			modifierGroups[idx].Modifiers = append(modifierGroups[idx].Modifiers, item)
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

func (j *JEClient) formDealProduct(r *http.Request, itemCode string, quantity int) (*Deal, error) {
	// Split itemCode into it's parts.
	itemCodes := strings.Split(itemCode, "|")
	// dealId := itemCodes[0]
	itemId := itemCodes[1]

	products := make(map[string]DealGroup)
	for items := range r.PostForm {
		if strings.Contains(items, "option") {
			// Extract the topping type and code
			var groupID string
			for i, s := range strings.Split(items, "[") {
				switch i {
				case 0:
					continue
				case 1:
					// Deal group id
					groupID = strings.Split(s, "]")[0]
				case 2:
					// Modifier Group ID
					// Can be possible to have duplicates of a modifier. In this case we want to increment the quantity.
					notModifiedId := strings.Split(s, "]")[0]
					if notModifiedId[len(notModifiedId)-2] == '_' {
						// We got one. We should also strip the suffix.
						notModifiedId = notModifiedId[:len(notModifiedId)-2]
					}

					key, err := j.GetKey(notModifiedId)
					if err != nil {
						return nil, err
					}

					key = strings.Split(key, "|")[0]
					if m, ok := products[groupID]; ok {
						m.Products = append(products[groupID].Products, Product{
							ProductId: key,
							Quantity:  1,
						})
					} else {
						products[groupID] = DealGroup{
							DealGroupId: groupID,
							Products: []Product{
								{
									ProductId: key,
									Quantity:  1,
								},
							},
						}
					}
				}
			}
		}
	}

	var dealGroups []DealGroup
	for _, group := range products {
		dealGroups = append(dealGroups, group)
	}

	return &Deal{
		Date:           time.Now().UTC().Format("2006-01-02T15:01:05.000Z"),
		ProductId:      itemId,
		Quantity:       quantity,
		ModifierGroups: nil,
		DealGroups:     dealGroups,
	}, nil
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

	itemCode := r.PostForm.Get("itemCode")
	itemCode, err = j.GetKey(itemCode)
	if err != nil {
		return "", err
	}

	quantityStr := r.PostForm.Get("quantity")

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		return "", err
	}

	var products []Product
	var deals []Deal
	itemCodes := strings.Split(itemCode, "|")
	if len(itemCodes) == 2 {
		deal, err := j.formDealProduct(r, itemCode, quantity)
		if err != nil {
			return "", err
		}

		deals = append(deals, *deal)
	} else {
		product, err := j.formProduct(r, itemCode, quantity)
		if err != nil {
			return "", err
		}

		products = append(products, *product)
	}

	basket := Basket{
		RestaurantSEOName: shopCode,
		MenuGroupId:       c,
		ServiceType:       "delivery",
		Products:          products,
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
		Deals:      deals,
		BasketMode: "None",
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/basket", j.KongAPIURL), basket)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

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
		Deals:      []Deal{},
		BasketMode: "None",
	}

	resp, err := j.httpPost(fmt.Sprintf("%s/basket", j.KongAPIURL), basket)
	if err != nil {
		return ""
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	var b struct {
		BasketId string `json:"BasketId"`
	}

	err = json.Unmarshal(body, &b)
	return b.BasketId
}

// buildAddEdit builds the Added half of a BasketEdit from the form, shared by EditBasket and ModifyBasketItem
func (j *JEClient) buildAddEdit(basketId string, r *http.Request) (BasketEdit, error) {
	itemCode, err := j.GetKey(r.PostForm.Get("itemCode"))
	if err != nil {
		return BasketEdit{}, err
	}

	quantity, err := strconv.Atoi(r.PostForm.Get("quantity"))
	if err != nil {
		return BasketEdit{}, err
	}

	edit := BasketEdit{BasketId: basketId}
	itemCodes := strings.Split(itemCode, "|")
	if len(itemCodes) == 3 {
		deal, err := j.formDealProduct(r, itemCode, quantity)
		if err != nil {
			return BasketEdit{}, err
		}

		edit.Deal.Added = []Deal{*deal}
	} else {
		product, err := j.formProduct(r, itemCode, quantity)
		if err != nil {
			return BasketEdit{}, err
		}

		edit.Product.Added = []Product{*product}
	}

	return edit, nil
}

func (j *JEClient) EditBasket(basketId string, r *http.Request) error {
	edit, err := j.buildAddEdit(basketId, r)
	if err != nil {
		return err
	}

	resp, err := j.httpPut(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId), edit)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrBasketEditFailed
	}

	return nil
}

// RemoveItem removes a line by its BasketProductIds since Just Eat no-ops a removal keyed by its ProductId
func (j *JEClient) RemoveItem(basketId string, ref BasketItemRef) error {
	if len(ref.BasketProductIds) == 0 {
		return ErrNoBasketProductIds
	}

	date := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	removed := make([]BasketRemoval, len(ref.BasketProductIds))
	for i, id := range ref.BasketProductIds {
		removed[i] = BasketRemoval{Date: date, BasketProductId: id}
	}

	edit := BasketEdit{BasketId: basketId}
	if ref.IsDeal {
		edit.Deal = BasketStatusDeal{Removed: removed}
	} else {
		edit.Product = BasketStatusProduct{Removed: removed}
	}

	resp, err := j.httpPut(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId), edit)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrBasketRemovalFailed
	}

	return nil
}

// ModifyBasketItem does two sequential PUTs since one PUT with both removed and added drops the added half
func (j *JEClient) ModifyBasketItem(basketId string, oldRef BasketItemRef, r *http.Request) error {
	if len(oldRef.BasketProductIds) == 0 {
		return ErrNoBasketProductIds
	}

	// build the replacement before removing anything so a failed modifier lookup leaves the old line intact
	edit, err := j.buildAddEdit(basketId, r)
	if err != nil {
		return err
	}

	if err := j.RemoveItem(basketId, oldRef); err != nil {
		return err
	}

	resp, err := j.httpPut(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId), edit)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrBasketEditFailed
	}

	return nil
}

// getBasket returns the basket object from Just Eat.
func (j *JEClient) getBasket(basketId string) (BasketData, error) {
	resp, err := j.httpGet(fmt.Sprintf("%s/basket/%s", j.KongAPIURL, basketId))
	if err != nil {
		return BasketData{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(_Basket, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return BasketData{}, err
	}

	var summary BasketData
	err = json.Unmarshal(body, &summary)
	return summary, err
}

func lookupCategoryID(categoryIDs map[string]string, productId string) string {
	if id, ok := categoryIDs[productId]; ok {
		return id
	}

	if idx := strings.LastIndex(productId, "-"); idx != -1 {
		if id, ok := categoryIDs[productId[:idx]]; ok {
			return id
		}
	}

	return ""
}

// GetBasket returns the basket in a Demae usable structure.
func (j *JEClient) GetBasket(basketId string, r *http.Request) ([]any, error) {
	summary, err := j.getBasket(basketId)
	if err != nil {
		return nil, err
	}

	// real category IDs let the Wii resolve menuCode back to a category on "Change"
	var categoryIDs map[string]string
	if r != nil && r.URL != nil && r.URL.Query().Get("shopCode") != "" {
		categoryIDs, err = j.GetProductCategoryIDs(r.URL.Query().Get("shopCode"))
		if err != nil && !errors.Is(err, ErrNoMenuAvailable) {
			return nil, err
		}
	}

	// itemIndex numbers lines across Products and Deals since basket_delete only echoes back this position
	itemIndex := 0

	var refs []BasketItemRef
	var basketItems []demae.BasketItem
	for _, product := range summary.BasketSummary.Products {
		// First group the modifiers
		var modifiers []any
		for i, option := range product.ModifierGroups {
			groupCode, err := j.ShortenID(option.GroupId)
			if err != nil {
				return nil, err
			}

			group := demae.ItemOne{
				XMLName: xml.Name{Local: fmt.Sprintf("container%d", i)},
				Info:    demae.CDATA{Value: ""},
				Code:    demae.CDATA{Value: groupCode},
				Type:    demae.CDATA{Value: 0},
				Name:    demae.CDATA{Value: fmt.Sprintf("Modifier %d", i+1)},
				List:    demae.KVFieldWChildren{},
			}

			for _, modifier := range option.Modifiers {
				modifierCode, err := j.ShortenID(modifier.ID)
				if err != nil {
					return nil, err
				}

				group.List.Value = append(group.List.Value, demae.Item{
					MenuCode:   demae.CDATA{Value: modifierCode},
					ItemCode:   demae.CDATA{Value: modifierCode},
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

		productCode, err := j.ShortenID(product.ProductId)
		if err != nil {
			return nil, err
		}

		refs = append(refs, BasketItemRef{IsDeal: false, BasketProductIds: product.BasketProductIds})

		priceStr := fmt.Sprintf("$%.2f", product.UnitPrice)
		amountStr := fmt.Sprintf("$%.2f", product.TotalPrice)
		basketItems = append(basketItems, demae.BasketItem{
			XMLName:       xml.Name{Local: fmt.Sprintf("container%d", itemIndex)},
			BasketNo:      demae.CDATA{Value: itemIndex},
			MenuCode:      demae.CDATA{Value: lookupCategoryID(categoryIDs, product.ProductId)},
			ItemCode:      demae.CDATA{Value: productCode},
			Name:          demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(product.Name), 26, -1)},
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
		itemIndex++
	}

	// Now we process any deal products.
	for _, product := range summary.BasketSummary.Deals {
		var modifiers []any
		for _, dealGroup := range product.DealGroups {
			for i, _product := range dealGroup.Products {
				dealProductCode, err := j.ShortenID(_product.ProductId)
				if err != nil {
					return nil, err
				}

				group := demae.ItemOne{
					XMLName: xml.Name{Local: fmt.Sprintf("container%d", i)},
					Info:    demae.CDATA{Value: ""},
					Code:    demae.CDATA{Value: dealProductCode},
					Type:    demae.CDATA{Value: 0},
					Name:    demae.CDATA{Value: _product.Name},
					List:    demae.KVFieldWChildren{},
				}

				for _, modifierGroup := range _product.ModifierGroups {
					for _, modifier := range modifierGroup.Modifiers {
						dealModifierCode, err := j.ShortenID(modifier.ID)
						if err != nil {
							return nil, err
						}

						group.List.Value = append(group.List.Value, demae.Item{
							MenuCode:   demae.CDATA{Value: dealModifierCode},
							ItemCode:   demae.CDATA{Value: dealModifierCode},
							Name:       demae.CDATA{Value: modifier.Name},
							Price:      demae.CDATA{Value: 0},
							Info:       demae.CDATA{Value: 0},
							IsSelected: &demae.CDATA{Value: demae.BoolToInt(true)},
							Image:      demae.CDATA{Value: 0},
							IsSoldout:  demae.CDATA{Value: demae.BoolToInt(false)},
						})
					}
				}

				if len(group.List.Value) == 0 {
					// Form with just the product
					group.List.Value = append(group.List.Value, demae.Item{
						MenuCode:   demae.CDATA{Value: dealProductCode},
						ItemCode:   demae.CDATA{Value: dealProductCode},
						Name:       demae.CDATA{Value: _product.Name},
						Price:      demae.CDATA{Value: _product.TotalPrice},
						Info:       demae.CDATA{Value: 0},
						IsSelected: &demae.CDATA{Value: demae.BoolToInt(true)},
						Image:      demae.CDATA{Value: 0},
						IsSoldout:  demae.CDATA{Value: demae.BoolToInt(false)},
					})
				}

				modifiers = append(modifiers, group)
			}
		}

		dealCode, err := j.ShortenID(product.ProductId)
		if err != nil {
			return nil, err
		}

		refs = append(refs, BasketItemRef{IsDeal: true, BasketProductIds: product.BasketProductIds})

		priceStr := fmt.Sprintf("$%.2f", product.UnitPrice)
		amountStr := fmt.Sprintf("$%.2f", product.TotalPrice)
		basketItems = append(basketItems, demae.BasketItem{
			XMLName:       xml.Name{Local: fmt.Sprintf("container%d", itemIndex)},
			BasketNo:      demae.CDATA{Value: itemIndex},
			MenuCode:      demae.CDATA{Value: lookupCategoryID(categoryIDs, product.ProductId)},
			ItemCode:      demae.CDATA{Value: dealCode},
			Name:          demae.CDATA{Value: demae.Wordwrap(demae.RemoveInvalidCharacters(product.Name), 26, -1)},
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
		itemIndex++
	}

	if err := j.SetBasketItems(basketId, refs); err != nil {
		return nil, err
	}

	basketPrice := demae.KVField{
		XMLName: xml.Name{Local: "basketPrice"},
		Value:   summary.BasketSummary.BasketTotals.Subtotal,
	}

	// Demae rounds anything less than 1 to 0. Round up to avoid this.
	deliveryCharge := summary.BasketSummary.DeliveryCharge
	if deliveryCharge < 1 {
		deliveryCharge = 1
	}

	chargePrice := demae.KVField{
		XMLName: xml.Name{Local: "chargePrice"},
		Value:   deliveryCharge,
	}

	totalPrice := demae.KVField{
		XMLName: xml.Name{Local: "totalPrice"},
		Value:   summary.BasketSummary.BasketTotals.Total,
	}

	discountPrice := demae.KVField{
		XMLName: xml.Name{Local: "discountPrice"},
		Value:   summary.BasketSummary.TotalDiscount,
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
		discountPrice,
		status,
		cart,
	}, nil
}
