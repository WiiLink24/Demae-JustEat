package server

import (
	"encoding/json"
	"net/http"

	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/gin-gonic/gin"
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

func clearOrder(hollywoodId string) error {
	_, err := pool.Exec(ctx, ClearOrder, hollywoodId)
	return err
}

func displayPaymentScreen(c *gin.Context) {
	_wiis, ok := c.Get("wiis")
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
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

	var activeOrdersArray []ActiveOrder
	for _, wii := range _wiis.([]Wii) {
		if basket, ok := activeOrders[uint32(wii.HollywoodID)]; ok {
			activeOrdersArray = append(activeOrdersArray, ActiveOrder{
				WiiNumber:   wii.WiiNumber,
				HollywoodID: uint32(wii.HollywoodID),
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
