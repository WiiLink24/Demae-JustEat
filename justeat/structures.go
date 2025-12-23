package justeat

type Category struct {
	Id           string        `json:"Id"`
	Name         string        `json:"Name"`
	Description  string        `json:"Description"`
	Preview      string        `json:"Preview"`
	ItemIds      []string      `json:"ItemIds"`
	ImageSources []interface{} `json:"ImageSources"`
}

type Menu struct {
	MenuGroupId  string   `json:"MenuGroupId"`
	Description  string   `json:"Description"`
	ServiceTypes []string `json:"ServiceTypes"`
	Schedules    []struct {
		DayOfWeek string `json:"DayOfWeek"`
		Times     []struct {
			FromLocalTime string `json:"FromLocalTime"`
			ToLocalTime   string `json:"ToLocalTime"`
		} `json:"Times"`
		OrderTimeSchedule []struct {
			LowerBound float64 `json:"LowerBound"`
		} `json:"OrderTimeSchedule"`
		BaseWorkingTimeMinutes int `json:"BaseWorkingTimeMinutes"`
	} `json:"Schedules"`
	Categories           []Category `json:"Categories"`
	PreOrderEnabledTimes []struct {
		DayOfWeek string `json:"DayOfWeek"`
		Times     []struct {
			FromLocalTime string `json:"FromLocalTime"`
			ToLocalTime   string `json:"ToLocalTime"`
		} `json:"Times"`
	} `json:"PreOrderEnabledTimes"`
	HasManyCategoryImages bool `json:"HasManyCategoryImages"`
	HasManyProductImages  bool `json:"HasManyProductImages"`
}

type Restaurant struct {
	RestaurantId   string `json:"RestaurantId"`
	RestaurantInfo struct {
		Name        string `json:"Name"`
		Description string `json:"Description"`
		IsOffline   bool   `json:"IsOffline"`
		PhoneNumber string `json:"AllergenPhoneNumber"`
		Location    struct {
			Address   string  `json:"Address"`
			Latitude  float64 `json:"Latitude"`
			Longitude float64 `json:"Longitude"`
		} `json:"Location"`
		RestaurantOpeningTimes []struct {
			ServiceType string `json:"ServiceType"`
			TimesPerDay []struct {
				DayOfWeek string `json:"DayOfWeek"`
				Times     []struct {
					FromLocalTime string `json:"FromLocalTime"`
					ToLocalTime   string `json:"ToLocalTime"`
				} `json:"Times"`
			} `json:"TimesPerDay"`
		} `json:"RestaurantOpeningTimes"`
	} `json:"RestaurantInfo"`
	Menus          []Menu `json:"Menus"`
	ItemsUrl       string `json:"ItemsUrl"`
	ItemDetailsUrl string `json:"ItemDetailsUrl"`
}

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id           string `json:"Id"`
	Name         string `json:"Name"`
	Description  string `json:"Description"`
	ImageSources []struct {
		Path string `json:"Path"`
	} `json:"ImageSources"`
	Type          string        `json:"Type"`
	Labels        []interface{} `json:"Labels"`
	Variations    []Variation   `json:"Variations"`
	EnergyContent struct {
		EnergyDisplay string `json:"EnergyDisplay"`
	} `json:"EnergyContent"`
	NumberOfServings struct {
		ServingsDisplay string `json:"ServingsDisplay"`
	} `json:"NumberOfServings"`
	HasVariablePrice         bool   `json:"HasVariablePrice"`
	HasVariableEnergyContent bool   `json:"HasVariableEnergyContent"`
	EnergyUnits              string `json:"EnergyUnits"`
	HasVariableServings      bool   `json:"HasVariableServings"`
}

type Variation struct {
	Id                string      `json:"Id"`
	Name              string      `json:"Name"`
	Type              string      `json:"Type"`
	BasePrice         float64     `json:"BasePrice"`
	DealOnly          bool        `json:"DealOnly"`
	MenuGroupIds      []string    `json:"MenuGroupIds"`
	ModifierGroupsIds []string    `json:"ModifierGroupsIds"`
	DealGroupsIds     []string    `json:"DealGroupsIds"`
	NutritionalInfo   interface{} `json:"NutritionalInfo"`
	NumberOfServings  interface{} `json:"NumberOfServings"`
}

type Modifiers struct {
	ModifierGroups []struct {
		Id         string   `json:"Id"`
		Name       string   `json:"Name"`
		MinChoices int      `json:"MinChoices"`
		MaxChoices int      `json:"MaxChoices"`
		Modifiers  []string `json:"Modifiers"`
	} `json:"ModifierGroups"`
	DealGroups []struct {
		Id                 string `json:"Id"`
		Name               string `json:"Name"`
		NumberOfChoices    int    `json:"NumberOfChoices"`
		DealItemVariations []struct {
			DealItemVariationId string      `json:"DealItemVariationId"`
			MinChoices          int         `json:"MinChoices"`
			MaxChoices          int         `json:"MaxChoices"`
			AdditionPrice       float64     `json:"AdditionPrice"`
			NumberOfServings    interface{} `json:"NumberOfServings"`
		} `json:"DealItemVariations"`
	} `json:"DealGroups"`
	ModifierSets []struct {
		Id       string `json:"Id"`
		Modifier struct {
			Id               string      `json:"Id"`
			Name             string      `json:"Name"`
			AdditionPrice    float64     `json:"AdditionPrice"`
			RemovePrice      float64     `json:"RemovePrice"`
			DefaultChoices   int         `json:"DefaultChoices"`
			MinChoices       int         `json:"MinChoices"`
			MaxChoices       int         `json:"MaxChoices"`
			NutritionalInfo  interface{} `json:"NutritionalInfo"`
			NumberOfServings interface{} `json:"NumberOfServings"`
		} `json:"Modifier"`
	} `json:"ModifierSets"`
}

type Modifier struct {
	ID       string `json:"modifierId"`
	Quantity int    `json:"quantity"`
}

type ModifierGroup struct {
	GroupId   string     `json:"modifierGroupId"`
	Modifiers []Modifier `json:"modifiers"`
}

type Product struct {
	Date               string          `json:"date"`
	ProductId          string          `json:"productId"`
	Quantity           int             `json:"quantity"`
	ModifierGroups     []ModifierGroup `json:"modifierGroups"`
	RemovedIngredients []any           `json:"removedIngredients"`
}

type Deal struct {
	Date           string          `json:"date"`
	ProductId      string          `json:"productId"`
	Quantity       int             `json:"quantity"`
	ModifierGroups []ModifierGroup `json:"modifierGroups"`
	DealGroups     []DealGroup     `json:"dealGroups"`
}

type DealGroup struct {
	DealGroupId string    `json:"dealGroupId"`
	Products    []Product `json:"products"`
}

type Location struct {
	ZipCode     string      `json:"zipCode"`
	GeoLocation GeoLocation `json:"geoLocation"`
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type OrderDetails struct {
	Location Location `json:"location"`
}

type Basket struct {
	RestaurantSEOName string       `json:"restaurantSeoName"`
	MenuGroupId       string       `json:"menuGroupId"`
	ServiceType       string       `json:"serviceType"`
	Products          []Product    `json:"products"`
	OrderDetails      OrderDetails `json:"orderDetails"`
	Deals             []Deal       `json:"deals"`
	Consents          []any        `json:"consents"`
	BasketMode        string       `json:"basketMode"`
}

type RequestedModifier struct {
	ID       string `json:"ModifierId"`
	Quantity int    `json:"Quantity"`
	Name     string `json:"Name"`
}

type RequestedModifierGroup struct {
	GroupId   string              `json:"ModifierGroupId"`
	Modifiers []RequestedModifier `json:"Modifiers"`
}

type RequestedBasket struct {
	Name           string                   `json:"Name"`
	UnitPrice      float64                  `json:"UnitPrice"`
	TotalPrice     float64                  `json:"TotalPrice"`
	ProductId      string                   `json:"ProductId"`
	Quantity       int                      `json:"Quantity"`
	ModifierGroups []RequestedModifierGroup `json:"ModifierGroups"`
}

type RequestedDeal struct {
	DealGroups []RequestedDealGroup `json:"DealGroups"`
	RequestedBasket
}

type RequestedDealGroup struct {
	Products []RequestedBasket `json:"Products"`
}

type BasketTotals struct {
	Subtotal float64 `json:"SubTotal"`
	Total    float64 `json:"Total"`
}

type BasketData struct {
	BasketSummary BasketSummary `json:"BasketSummary"`
}

type BasketAdjustment struct {
	Name       string `json:"name"`
	Adjustment any    `json:"Adjustment" json:"adjustment"`
}

type BasketSummary struct {
	Products       []RequestedBasket  `json:"Products"`
	Deals          []RequestedDeal    `json:"Deals"`
	BasketTotals   BasketTotals       `json:"BasketTotals"`
	Adjustments    []BasketAdjustment `json:"Adjustments"`
	DeliveryCharge float64            `json:"DeliveryCharge"`
	TotalDiscount  float64            `json:"TotalDiscount"`
}

type BasketRemoval struct {
	Date            string `json:"Date"`
	BasketProductId string `json:"BasketProductId"`
}

type BasketStatusProduct struct {
	Added   []Product       `json:"Added"`
	Updated []Product       `json:"Updated"`
	Removed []BasketRemoval `json:"Removed"`
}

type BasketStatusDeal struct {
	Added   []Deal          `json:"Added"`
	Updated []Product       `json:"Updated"`
	Removed []BasketRemoval `json:"Removed"`
}

type BasketEdit struct {
	BasketId string              `json:"BasketId"`
	Product  BasketStatusProduct `json:"Product"`
	Deal     BasketStatusDeal    `json:"Deal"`
}

type LineItem struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Quantity   string `json:"quantity"`
	UnitAmount string `json:"unit_amount"`
}

// WebBasket is a simplified and complete format of Basket which is used in our front facing web server.
type WebBasket struct {
	BasketURL       string     `json:"basket_url"`
	Items           []LineItem `json:"items"`
	Total           float64    `json:"total"`
	CurrencyISOCode string     `json:"currency_iso_code"`
}
