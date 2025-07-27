package justeat

import "time"

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

type Availability struct {
	Times []AvailabilityTimes `json:"times"`
	ASAP  bool                `json:"asapAvailable"`
}

type AvailabilityTimes struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type CheckoutTime struct {
	TimeFrom string `json:"timeFrom"`
	TimeTo   string `json:"timeTo"`
}

type CheckoutUser struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	// nil
	DateOfBirth *string `json:"dateOfBirth"`
}

type CheckoutAddress struct {
	Lines              []string `json:"lines"`
	Locality           string   `json:"locality"`
	AdministrativeArea *string  `json:"administrativeArea"`
	PostalCode         string   `json:"postalCode"`
}

type CheckoutLocation struct {
	Address     CheckoutAddress `json:"address"`
	GeoLocation GeoLocation     `json:"geolocation"`
}

type CheckoutFulfilment struct {
	Time struct {
		Asap      bool         `json:"asap"`
		Scheduled CheckoutTime `json:"scheduled"`
	} `json:"time"`
	Location CheckoutLocation `json:"location"`
	Table    *string          `json:"table"`
}

type Note struct {
	Note string `json:"note"`
}

type CheckoutNotes struct {
	Order   Note `json:"order"`
	Courier Note `json:"courier"`
	Kitchen Note `json:"kitchen"`
}

type CheckoutTipping struct {
	Courier struct {
		FixedAmount  *int `json:"fixedAmount"`
		CustomAmount *int `json:"customAmount"`
	} `json:"courier"`
}

type CheckoutPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type PaymentTypes struct {
	AvailablePaymentTypes []PaymentOption `json:"availablePaymentTypes"`
}

type PaymentOption struct {
	PaymentType    string                `json:"paymentType"`
	Status         string                `json:"status"`
	AdditionalData *AdditionalPaypalData `json:"additionalData"`
}

type AdditionalPaypalData struct {
	ClientKey string `json:"clientKey"`
}

type BrainTreeConfig struct {
	AuthFingerprint string `json:"authorizationFingerprint"`
	MerchantID      string `json:"merchantId"`
	ConfigURL       string `json:"configUrl"`
	ClientAPIUrl    string `json:"clientApiUrl"`
	Paypal          struct {
		CurrencyCode string `json:"currencyIsoCode"`
	} `json:"paypal"`
}

type BrainTreeItem struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Quantity   string `json:"quantity"`
	UnitAmount string `json:"unit_amount"`
}

type BrainTreeExperience struct {
	NoShipping      bool   `json:"no_shipping"`
	BrandName       string `json:"brand_name"`
	LocaleCode      string `json:"locale_code"`
	UserAction      string `json:"user_action"`
	AddressOverride bool   `json:"address_override"`
}

type BrainTreeCreatePaypal struct {
	ReturnURL                 string              `json:"return_url"`
	CancelURL                 string              `json:"cancel_url"`
	OfferPayLater             bool                `json:"offer_pay_later"`
	AuthorizationFingerprint  string              `json:"authorization_fingerprint"`
	Amount                    string              `json:"amount"`
	CurrencyISOCode           string              `json:"currency_iso_code"`
	Intent                    string              `json:"intent"`
	LineItems                 []BrainTreeItem     `json:"line_items"`
	Line1                     string              `json:"line1"`
	City                      string              `json:"city"`
	PostalCode                string              `json:"postal_code"`
	CountryCode               string              `json:"country_code"`
	RecipientName             string              `json:"recipient_name"`
	ExperienceProfile         BrainTreeExperience `json:"experience_profile"`
	AuthorizationFingerprint2 string              `json:"authorizationFingerprint"`
}

type BrainTreePaymentResourceHead struct {
	PaymentResource BrainTreePaymentResource `json:"paymentResource"`
}

type BrainTreePaymentResource struct {
	PaymentToken string `json:"paymentToken"`
	RedirectURL  string `json:"redirectUrl"`
}

type PaypalMetadata struct {
	AppGUID             string `json:"app_guid"`
	AppID               string `json:"app_id"`
	AndroidID           string `json:"android_id"`
	AppVersion          string `json:"app_version"`
	AppFirstInstallTime int    `json:"app_first_install_time"`
	AppLastUpdateTime   int    `json:"app_last_update_time"`
	ConfURL             string `json:"conf_url"`
	CompVersion         string `json:"comp_version"`
	DeviceModel         string `json:"device_model"`
	DeviceName          string `json:"device_name"`
	GSFID               string `json:"gsf_id"`
	IsEmulator          bool   `json:"is_emulator"`
	EF                  string `json:"ef"`
	IsRooted            bool   `json:"is_rooted"`
	RF                  string `json:"rf"`
	OSType              string `json:"os_type"`
	OSVersion           string `json:"os_version"`
	PayloadType         string `json:"payload_type"`
	SMSEnabled          bool   `json:"sms_enabled"`
	MagnesGUID          struct {
		ID        string `json:"id"`
		CreatedAt int    `json:"created_at"`
	} `json:"magnes_guid"`
	MagnesSource      int      `json:"magnes_source"`
	SourceAppVersion  string   `json:"source_app_version"`
	TotalStorageSpace int      `json:"total_storage_space"`
	T                 bool     `json:"t"`
	PairingID         string   `json:"pairing_id"`
	ConnType          string   `json:"conn_type"`
	ConfVersion       string   `json:"conf_version"`
	DMO               bool     `json:"dmo"`
	DCID              string   `json:"dc_id"`
	DeviceUptime      int      `json:"device_uptime"`
	IpAddrs           string   `json:"ip_addrs"`
	IpAddresses       []string `json:"ip_addresses"`
	LocaleCountry     string   `json:"locale_country"`
	LocaleLang        string   `json:"locale_lang"`
	PhoneType         string   `json:"phone_type"`
	RiskCompSessionID string   `json:"risk_comp_session_id"`
	Roaming           bool     `json:"roaming"`
	SimOperatorName   string   `json:"sim_operator_name"`
	Timestamp         int      `json:"timestamp"`
	TZName            string   `json:"tz_name"`
	DS                bool     `json:"ds"`
	TZ                int      `json:"tz"`
	NetworkOperator   string   `json:"network_operator"`
	ProxySetting      string   `json:"proxy_setting"`
	MGID              string   `json:"mg_id"`
	PL                string   `json:"pl"`
	SR                struct {
		AC bool `json:"ac"`
		GY bool `json:"gy"`
		MG bool `json:"mg"`
	} `json:"sr"`
}

type CombinedBrainTree struct {
	OrderID         string                       `json:"orderId"`
	Country         Country                      `json:"country"`
	Head            BrainTreePaymentResourceHead `json:"head"`
	BrainTree       BrainTreeCreatePaypal        `json:"brainTree"`
	Metadata        PaypalMetadata               `json:"metadata"`
	BrainTreeConfig BrainTreeConfig              `json:"brain_tree_config"`
}
