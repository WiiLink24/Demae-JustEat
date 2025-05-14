package justeat

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (j *JEClient) getLocalizedTimeLocation() (*time.Location, error) {
	return time.LoadLocation(timeZones[j.Country])
}

func (j *JEClient) getLocalizedTime(value string) (time.Time, error) {
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return time.Time{}, err
	}

	_t, err := time.Parse("2006-01-02T15:04:05Z", value)
	if err != nil {
		return time.Time{}, err
	}

	return _t.In(zone), nil
}

func (j *JEClient) getAvailableTimes(basketId string) (map[string]any, error) {
	resp, err := j.httpGet(fmt.Sprintf("%s/checkout/%s/%s/fulfilment/availabletimes", j.KongAPIURL, strings.ToLower(string(j.Country)), basketId))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var availability map[string]any
	err = json.Unmarshal(body, &availability)
	return availability, err
}

func (j *JEClient) GetAvailableTimes(basketId string) ([]demae.KVFieldWChildren, error) {
	availability, err := j.getAvailableTimes(basketId)
	if err != nil {
		return nil, err
	}

	var times []demae.KVFieldWChildren
	for i, _time := range availability["times"].([]any) {
		_t, err := j.getLocalizedTime(_time.(map[string]any)["from"].(string))
		if err != nil {
			return nil, err
		}

		times = append(times, demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "option"},
			Value: []any{
				demae.KVField{
					XMLName: xml.Name{Local: "id"},
					Value:   i,
				},
				demae.KVField{
					XMLName: xml.Name{Local: "name"},
					Value:   _t.Format("2006-01-02 15:04:05"),
				},
			},
		})
	}

	return times, err
}

func (j *JEClient) PlaceOrder(r *http.Request, basketId string) {
	storeId := r.PostForm.Get("shop[ShopCode]")
	firstName := r.PostForm.Get("member[Name1]")
	lastName := r.PostForm.Get("member[Name2]")
	phoneNumber := r.PostForm.Get("member[TelNo]")

	long, lat, city, err := j.getGeocodedAddress()
	if err != nil {
		log.Println(err)
	}

	user := CheckoutUser{
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
		DateOfBirth: nil,
	}

	location := CheckoutLocation{
		Address: CheckoutAddress{
			Lines:              []string{r.PostForm.Get("member[Address5]")},
			Locality:           city,
			AdministrativeArea: nil,
			PostalCode:         j.PostalCode,
		},
		GeoLocation: GeoLocation{
			Latitude:  lat,
			Longitude: long,
		},
	}

	// TODO: Allow for choosing a scheduled time
	times, err := j.getAvailableTimes(basketId)
	if err != nil {
		log.Println(err)
	}

	checkout := CheckoutFulfilment{
		Time: struct {
			Asap      bool         `json:"asap"`
			Scheduled CheckoutTime `json:"scheduled"`
		}{
			Asap: times["asapAvailable"].(bool),
			Scheduled: CheckoutTime{
				// The first available time
				// TODO: Scheduled times!!!!
				TimeFrom: times["times"].([]any)[0].(map[string]any)["from"].(string),
				TimeTo:   times["times"].([]any)[0].(map[string]any)["to"].(string),
			},
		},
		Location: location,
		Table:    nil,
	}

	notes := CheckoutNotes{}
	// I feel really evil for this, no tip
	tips := CheckoutTipping{}

	customerPatch := CheckoutPatch{
		Op:    "add",
		Path:  "/customer",
		Value: user,
	}

	fulfilmentPatch := CheckoutPatch{
		Op:    "add",
		Path:  "/fulfilment",
		Value: checkout,
	}

	notesPatch := CheckoutPatch{
		Op:    "add",
		Path:  "/notes",
		Value: notes,
	}

	tippingPatch := CheckoutPatch{
		Op:    "add",
		Path:  "/tipping",
		Value: tips,
	}

	total, err := j.prepareCheckout(basketId, customerPatch, fulfilmentPatch, notesPatch, tippingPatch)
	if err != nil {
		log.Println(err)
	}

	config, err := j.getPayPalToken(storeId, basketId)
	if err != nil {
		log.Println(err)
	}

	orderId, err := j.getOrderID(basketId, total, config.Paypal.CurrencyCode)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Order ID: " + orderId)
	// We have to get the entire basket.
	basket, err := j.getBasket(basketId)
	if err != nil {
		log.Println(err)
	}

	var receiptItems []BrainTreeItem
	for _, item := range basket.BasketSummary.Products {
		receiptItems = append(receiptItems, BrainTreeItem{
			Kind:       "debit",
			Name:       item.Name,
			Quantity:   strconv.Itoa(item.Quantity),
			UnitAmount: demae.FloatToString(item.UnitPrice),
		})
	}

	// Per giustino:
	// !!!very important step!!! failing to do this correctly it will error out at makePaypalURL as it verifies the amount and charge kind
	for _, adjustment := range basket.BasketSummary.Adjustments {
		adj := BrainTreeItem{
			Kind:       "debit",
			Name:       adjustment.Name,
			Quantity:   "1",
			UnitAmount: "",
		}

		if s, ok := adjustment.Adjustment.(string); ok {
			adj.Kind = "credit"
			adj.UnitAmount = strings.Replace(s, "-", "", -1)
		} else {
			adj.UnitAmount = demae.FloatToString(adjustment.Adjustment.(float64))
		}

		receiptItems = append(receiptItems, adj)
	}

	if basket.BasketSummary.DeliveryCharge != 0 {
		receiptItems = append(receiptItems, BrainTreeItem{
			Kind:       "debit",
			Name:       "Delivery",
			Quantity:   "1",
			UnitAmount: demae.FloatToString(basket.BasketSummary.DeliveryCharge),
		})
	}

	// Get the payment URL.
	brainTree := BrainTreeCreatePaypal{
		ReturnURL:                "customer-details-oneapp.braintree://onetouch/v1/success",
		CancelURL:                "customer-details-oneapp.braintree://onetouch/v1/cancel",
		OfferPayLater:            false,
		AuthorizationFingerprint: config.AuthFingerprint,
		Amount:                   demae.FloatToString(basket.BasketSummary.BasketTotals.Total),
		CurrencyISOCode:          config.Paypal.CurrencyCode,
		Intent:                   "authorize",
		LineItems:                receiptItems,
		Line1:                    r.PostForm.Get("member[Address5]"),
		City:                     city,
		PostalCode:               j.PostalCode,
		CountryCode:              string(j.Country),
		RecipientName:            firstName + " " + lastName,
		ExperienceProfile: BrainTreeExperience{
			NoShipping:      true,
			BrandName:       "Just Eat",
			LocaleCode:      languageCodes[j.Country],
			UserAction:      "commit",
			AddressOverride: true,
		},
		AuthorizationFingerprint2: config.AuthFingerprint,
	}

	if j.Country == UnitedKingdom {
		brainTree.CountryCode = "GB"
	}

	head, err := j.makePaypalURL(config, brainTree)
	if err != nil {
		log.Println(err)
	}

	meta, err := j.sendPaypalMetadata()
	if err != nil {
		log.Println(err)
	}

	combined := CombinedBrainTree{
		Head:            *head,
		BrainTree:       brainTree,
		Metadata:        *meta,
		BrainTreeConfig: *config,
	}

	// Encode as JSON
	combinedBytes, err := json.Marshal(combined)
	if err != nil {
		log.Println(err)
	}

	_, err = j.db.Exec(j.Context, UpdateBraintree, string(combinedBytes), j.WiiID)
	if err != nil {
		log.Println(err)
	}
	// Remaining parts of the process have to be completed post order.
	// TODO: Figure out how to get the user the PayPal payment URL.
}

func (j *JEClient) prepareCheckout(basketId string, patches ...CheckoutPatch) (totalCost int, err error) {
	_url := fmt.Sprintf("%s/checkout/%s/%s", j.KongAPIURL, strings.ToLower(string(j.Country)), basketId)
	resp, err := j.httpPatch(_url, patches)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	var payload map[string]any
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return 0, err
	}

	if !payload["isFulfillable"].(bool) {
		return 0, NotFulfillable
	}

	return int(payload["purchase"].(map[string]any)["total"].(map[string]any)["price"].(map[string]any)["amount"].(float64)), err
}
