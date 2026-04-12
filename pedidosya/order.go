package justeat

import (
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"net/http"
	"strconv"
	"strings"
)

func (j *JEClient) PlaceOrder(r *http.Request, basketId string) error {
	basket, err := j.getBasket(basketId)
	if err != nil {
		return err
	}

	paymentTotal := 0.0
	var receiptItems []LineItem
	for _, item := range basket.BasketSummary.Products {
		receiptItems = append(receiptItems, LineItem{
			Kind:       "debit",
			Name:       item.Name,
			Quantity:   strconv.Itoa(item.Quantity),
			UnitAmount: demae.FloatToString(item.UnitPrice),
		})

		paymentTotal += item.UnitPrice
	}

	// Deals
	for _, deal := range basket.BasketSummary.Deals {
		receiptItems = append(receiptItems, LineItem{
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
		adj := LineItem{
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
		receiptItems = append(receiptItems, LineItem{
			Kind:       "debit",
			Name:       "Delivery",
			Quantity:   "1",
			UnitAmount: demae.FloatToString(basket.BasketSummary.DeliveryCharge),
		})

		paymentTotal += basket.BasketSummary.DeliveryCharge
	}

	// Encode as JSON
	combinedBytes, err := json.Marshal(WebBasket{
		BasketURL:       fmt.Sprintf("%s/checkout?basket=%s", BasketURLs[j.Country], basketId),
		Items:           receiptItems,
		Total:           paymentTotal,
		CurrencyISOCode: CurrencyISOCodes[j.Country],
	})
	if err != nil {
		return err
	}

	_, err = j.Db.Exec(j.Context, UpdateBasket, string(combinedBytes), j.WiiID)
	if err != nil {
		return err
	}

	return nil
}
