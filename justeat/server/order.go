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
	Basket      justeat.WebBasket
}

const (
	GetOrderParams = `SELECT basket, wii_id FROM users WHERE email = $1 AND basket IS NOT NULL`
	GetOrderForWii = `SELECT basket FROM users WHERE wii_id = $1 AND basket IS NOT NULL`
	ClearOrder     = `UPDATE users SET basket = NULL, basket_id = NULL WHERE wii_id = $1`
)

func getActiveOrders(email string) (map[uint32]justeat.WebBasket, error) {
	rows, err := pool.Query(ctx, GetOrderParams, email)
	if err != nil {
		return nil, err
	}

	payloads := map[uint32]justeat.WebBasket{}

	defer rows.Close()
	for rows.Next() {
		var data string
		var wiiId uint32
		err = rows.Scan(&data, &wiiId)
		if err != nil {
			return nil, err
		}

		var payload justeat.WebBasket
		err = json.Unmarshal([]byte(data), &payload)
		if err != nil {
			return nil, err
		}

		payloads[wiiId] = payload
	}

	return payloads, nil
}

func getActiveOrderForWii(hollywoodId string) (*justeat.WebBasket, error) {
	var data string
	err := pool.QueryRow(ctx, GetOrderForWii, hollywoodId).Scan(&data)
	if err != nil {
		return nil, err
	}

	var payload justeat.WebBasket
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
		if basket, ok := activeOrders[hollywoodID]; ok {
			activeOrdersArray = append(activeOrdersArray, ActiveOrder{
				WiiNumber:   wiiNoStr,
				HollywoodID: hollywoodID,
				Basket:      basket,
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

	err := clearOrder(hollywoodID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
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
