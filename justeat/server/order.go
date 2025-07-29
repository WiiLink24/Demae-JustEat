package server

import (
	"encoding/json"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/WiiLink24/nwc24"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ActiveOrder struct {
	WiiNumber   string
	HollywoodID uint32
	Braintree   justeat.CombinedBrainTree
}

const (
	GetOrderParams = `SELECT braintree, wii_id FROM users WHERE email = $1 AND braintree IS NOT NULL`
	GetOrderForWii = `SELECT braintree FROM users WHERE wii_id = $1 AND braintree IS NOT NULL`
	ClearOrder     = `UPDATE users SET braintree = NULL, basket_id = NULL WHERE wii_id = $1`
)

func getActiveOrders(email string) (map[uint32]justeat.CombinedBrainTree, error) {
	rows, err := pool.Query(ctx, GetOrderParams, email)
	if err != nil {
		return nil, err
	}

	payloads := map[uint32]justeat.CombinedBrainTree{}

	defer rows.Close()
	for rows.Next() {
		var data string
		var wiiId uint32
		err = rows.Scan(&data, &wiiId)
		if err != nil {
			return nil, err
		}

		var payload justeat.CombinedBrainTree
		err = json.Unmarshal([]byte(data), &payload)
		if err != nil {
			return nil, err
		}

		payloads[wiiId] = payload
	}

	return payloads, nil
}

func getActiveOrderForWii(hollywoodId string) (*justeat.CombinedBrainTree, error) {
	var data string
	err := pool.QueryRow(ctx, GetOrderForWii, hollywoodId).Scan(&data)
	if err != nil {
		return nil, err
	}

	var payload justeat.CombinedBrainTree
	err = json.Unmarshal([]byte(data), &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func clearOrder(hollywoodId string) error {
	_, err := pool.Exec(ctx, ClearOrder, hollywoodId)
	return err
}

func displayPaymentScreen(c *gin.Context) {
	justEatWiis, ok := c.Get("just_eat")
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	var linkedWiis []string
	for wiiNo, isLinked := range justEatWiis.(map[string]bool) {
		if !isLinked {
			continue
		}

		linkedWiis = append(linkedWiis, wiiNo)
	}

	email, ok := c.Get("email")
	if !ok {
		// How did we get here
		c.Status(http.StatusBadRequest)
		return
	}

	activeOrders, err := getActiveOrders(email.(string))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// This was a design oversight by me, the postgres database has the Hollywood ID while Authentik DB has the Wii Number.
	// We need both so we can allow the user to choose which Wii they want to for a purchase if for some reason they have more
	// than one active order, and post to the finalizePayment endpoint.
	var activeOrdersArray []ActiveOrder
	for _, wiiNoStr := range linkedWiis {
		// Convert to Wii ID and find active order if any.
		wiiNoInt, err := strconv.ParseUint(wiiNoStr, 10, 64)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		wiiNo := nwc24.LoadWiiNumber(wiiNoInt)
		hollywoodID := wiiNo.GetHollywoodID()
		if braintree, ok := activeOrders[hollywoodID]; ok {
			activeOrdersArray = append(activeOrdersArray, ActiveOrder{
				WiiNumber:   wiiNoStr,
				HollywoodID: hollywoodID,
				Braintree:   braintree,
			})
		}
	}

	c.HTML(http.StatusOK, "pay.html", gin.H{
		"ActiveOrders": activeOrdersArray,
	})
}

func finalizePayment(c *gin.Context) {
	hollywoodID := c.PostForm("hollywood_id")
	if hollywoodID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	order, err := getActiveOrderForWii(hollywoodID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Pseudo-client
	client := justeat.JEClient{
		Context:      ctx,
		Country:      order.Country,
		KongAPIURL:   justeat.KongAPIURLs[order.Country],
		GlobalAPIURL: justeat.GlobalMenuCDNURLs[order.Country],
		CheckoutURL:  justeat.CheckoutURLs[order.Country],
		Db:           pool,
	}

	err = client.SetAuth()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	retFirst := justeat.MakePaypalReturnURLFirst(order.Head)
	nonce, email, payerID, err := client.GetPaypalNonce(order.BrainTreeConfig, order.Metadata, order.BrainTree.AuthorizationFingerprint, retFirst)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	ret := justeat.MakePaypalReturnURL(order.Head.PaymentResource.PaymentToken, payerID)

	nonce, email, payerID, err = client.GetPaypalNonce(order.BrainTreeConfig, order.Metadata, order.BrainTree.AuthorizationFingerprint, ret)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	err = client.SendPayment(order.Metadata, nonce, email, payerID, order.OrderID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}

func cancelPayment(c *gin.Context) {
	hollywoodID := c.PostForm("hollywood_id")
	if hollywoodID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	err := clearOrder(hollywoodID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
}
