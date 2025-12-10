package justeat

import (
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/logger"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const OrderModule = "ORDER"

func (j *JEClient) PlaceOrder(r *http.Request, basketId string) error {
	storeId := r.PostForm.Get("shop[ShopCode]")
	firstName := r.PostForm.Get("member[Name1]")
	lastName := r.PostForm.Get("member[Name2]")
	phoneNumber := r.PostForm.Get("member[TelNo]")

	long, lat, city, err := j.getGeocodedAddress()
	if err != nil {
		return err
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
		return err
	}

	checkout := CheckoutFulfilment{
		Time: struct {
			Asap      bool         `json:"asap"`
			Scheduled CheckoutTime `json:"scheduled"`
		}{
			Asap: times["asapAvailable"].(bool),
			Scheduled: CheckoutTime{
				// The first available time
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
		return err
	}

	config, err := j.getPayPalToken(storeId, basketId)
	if err != nil {
		return err
	}

	orderId, err := j.getOrderID(basketId, total, config.Paypal.CurrencyCode)
	if err != nil {
		return err
	}

	// We have to get the entire basket.
	basket, err := j.getBasket(basketId)
	if err != nil {
		return err
	}

	// Contrary to the basket's total field, it is not actually the total.
	// We have to manually sum all products and fees.
	paymentTotal := 0.0
	var receiptItems []BrainTreeItem
	for _, item := range basket.BasketSummary.Products {
		receiptItems = append(receiptItems, BrainTreeItem{
			Kind:       "debit",
			Name:       item.Name,
			Quantity:   strconv.Itoa(item.Quantity),
			UnitAmount: demae.FloatToString(item.UnitPrice),
		})

		paymentTotal += item.UnitPrice
	}

	// Deals
	for _, deal := range basket.BasketSummary.Deals {
		receiptItems = append(receiptItems, BrainTreeItem{
			Kind:       "debit",
			Name:       deal.Name,
			Quantity:   strconv.Itoa(deal.Quantity),
			UnitAmount: demae.FloatToString(deal.UnitPrice),
		})

		paymentTotal += deal.UnitPrice
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
			paymentTotal += adjustment.Adjustment.(float64)
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

		paymentTotal += basket.BasketSummary.DeliveryCharge
	}

	// Get the payment URL.
	brainTree := BrainTreeCreatePaypal{
		ReturnURL:                "customer-details-oneapp.braintree://onetouch/v1/success",
		CancelURL:                "customer-details-oneapp.braintree://onetouch/v1/cancel",
		OfferPayLater:            false,
		AuthorizationFingerprint: config.AuthFingerprint,
		Amount:                   demae.FloatToString(paymentTotal),
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
			LocaleCode:      LanguageCodes[j.Country],
			UserAction:      "commit",
			AddressOverride: true,
		},
		AuthorizationFingerprint2: config.AuthFingerprint,
	}

	// United Kingdom edge case
	if j.Country == UnitedKingdom {
		brainTree.CountryCode = "GB"
	}

	head, err := j.makePaypalURL(config, brainTree)
	if err != nil {
		return err
	}

	meta, err := j.sendPaypalMetadata()
	if err != nil {
		return err
	}

	combined := CombinedBrainTree{
		OrderID:         orderId,
		Country:         j.Country,
		Head:            *head,
		BrainTree:       brainTree,
		Metadata:        *meta,
		BrainTreeConfig: *config,
	}

	// Encode as JSON
	combinedBytes, err := json.Marshal(combined)
	if err != nil {
		return err
	}

	_, err = j.Db.Exec(j.Context, UpdateBraintree, string(combinedBytes), j.WiiID)
	if err != nil {
		return err
	}

	// Remaining parts of the process have to be completed post order.
	return nil
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
		logger.Debug(OrderModule, payload)
		return 0, NotFulfillable
	}

	return int(payload["purchase"].(map[string]any)["total"].(map[string]any)["price"].(map[string]any)["amount"].(float64)), err
}
