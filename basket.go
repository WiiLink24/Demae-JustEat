package main

import (
	"context"
	"encoding/xml"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/gofrs/uuid"
	"log"
	"time"
)

const (
	DoesAuthKeyExist = `SELECT EXISTS(SELECT 1 FROM users WHERE users.wii_id = $1 AND users.auth_key IS NOT NULL)`
	InsertAuthkey    = `UPDATE users SET auth_key = $1 WHERE wii_id = $2`
	ClearBasket      = `UPDATE users SET basket_id = NULL WHERE wii_id = $1`
	InsertBasketID   = `UPDATE users SET basket_id = $1 WHERE wii_id = $2`
	DoesBasketExist  = `SELECT EXISTS(SELECT 1 FROM users WHERE users.wii_id = $1 AND users.basket_id IS NOT NULL)`
	GetBasketID      = `SELECT basket_id FROM users WHERE wii_id = $1`
)

func authKey(r *Response) {
	authKeyValue, err := uuid.DefaultGenerator.NewV1()
	if err != nil {
		r.ReportError(err)
		return
	}

	// First we query to determine if the user already has an auth key. If they do, reset the basket.
	var authExists bool
	row := pool.QueryRow(context.Background(), DoesAuthKeyExist, r.GetHollywoodId())
	err = row.Scan(&authExists)
	if err != nil {
		r.ReportError(err)
		return
	}

	if authExists {
		_, err = pool.Exec(context.Background(), ClearBasket, r.GetHollywoodId())
		if err != nil {
			r.ReportError(err)
			return
		}
	}

	_, err = pool.Exec(context.Background(), InsertAuthkey, authKeyValue.String(), r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = []any{
		demae.KVField{
			XMLName: xml.Name{Local: "authKey"},
			Value:   authKeyValue.String(),
		},
	}
}

func basketAdd(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	// Determine if we have a basket
	var basketExists bool
	row := pool.QueryRow(context.Background(), DoesBasketExist, r.GetHollywoodId())
	err = row.Scan(&basketExists)
	if err != nil {
		r.ReportError(err)
		return
	}

	if basketExists {
		// Edit basket
		var basketId string
		err = pool.QueryRow(context.Background(), GetBasketID, r.GetHollywoodId()).Scan(&basketId)
		if err != nil {
			r.ReportError(err)
			return
		}

		err = client.EditBasket(basketId, r.request)
		if err != nil {
			r.ReportError(err)
			return
		}
	} else {
		// Create basket
		basketId, err := client.CreateBasket(r.request)
		if err != nil {
			r.ReportError(err)
			return
		}

		_, err = pool.Exec(context.Background(), InsertBasketID, basketId, r.GetHollywoodId())
		if err != nil {
			r.ReportError(err)
			return
		}
	}
}

func basketList(r *Response) {
	var basketId string
	err := pool.QueryRow(context.Background(), GetBasketID, r.GetHollywoodId()).Scan(&basketId)
	if err != nil {
		r.ReportError(err)
		return
	}

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	basket, err := client.GetBasket(basketId, r.request)
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = basket
}

func basketReset(r *Response) {
	_, err := pool.Exec(context.Background(), ClearBasket, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}
}

func basketDelete(r *Response) {
	var basketId string
	err := pool.QueryRow(context.Background(), GetBasketID, r.GetHollywoodId()).Scan(&basketId)
	if err != nil {
		r.ReportError(err)
		return
	}

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	// Actually the product ID
	productID := r.request.URL.Query().Get("basketNo")
	err = client.RemoveItem(basketId, productID, r.request)
	if err != nil {
		r.ReportError(err)
	}
}

func orderDone(r *Response) {
	var basketId string
	err := pool.QueryRow(context.Background(), GetBasketID, r.GetHollywoodId()).Scan(&basketId)
	if err != nil {
		r.ReportError(err)
		return
	}

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	err = client.PlaceOrder(r.request, basketId)
	if err != nil {
		r.ReportError(err)
		log.Println(err)
		return
	}

	currentTime := time.Now().Format("200602011504")
	r.AddKVWChildNode("Message", demae.KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "Thank you! Your order has been placed!",
	})
	r.AddKVNode("order_id", "1")
	r.AddKVNode("orderDay", currentTime)
	r.AddKVNode("hashKey", "Testing: 1, 2, 3")
	r.AddKVNode("hour", currentTime)
}
