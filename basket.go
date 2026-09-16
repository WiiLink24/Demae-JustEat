package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
)

const (
	DoesAuthKeyExist = `SELECT EXISTS(SELECT 1 FROM users WHERE users.wii_id = $1 AND users.auth_key IS NOT NULL)`
	InsertAuthkey    = `UPDATE users SET auth_key = $1 WHERE wii_id = $2`
	ClearBasket      = `UPDATE users SET basket_id = NULL WHERE wii_id = $1`
	InsertBasketID   = `UPDATE users SET basket_id = $1 WHERE wii_id = $2`
	// COALESCE avoids a NULL scan panic when a Wii has no basket
	GetBasketID = `SELECT COALESCE(basket_id, '') FROM users WHERE wii_id = $1`
)

// ErrNoBasket means the Wii has no basket right now (never created, reset, or emptied)
var ErrNoBasket = demae.NewSentryError("Your basket is empty. Please add an\nitem first.", false)

// getBasketID wraps the NULL case as ErrNoBasket instead of a raw NULL scan error
func getBasketID(hollywoodID string) (string, error) {
	var basketId string
	if err := pool.QueryRow(context.Background(), GetBasketID, hollywoodID).Scan(&basketId); err != nil {
		return "", err
	}

	if basketId == "" {
		return "", ErrNoBasket
	}

	return basketId, nil
}

func authKey(r *Response) {
	authKeyValue := demae.UUID()

	// First we query to determine if the user already has an auth key. If they do, reset the basket.
	var authExists bool
	row := pool.QueryRow(context.Background(), DoesAuthKeyExist, r.GetHollywoodId())
	err := row.Scan(&authExists)
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

	_, err = pool.Exec(context.Background(), InsertAuthkey, authKeyValue, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = []any{
		demae.KVField{
			XMLName: xml.Name{Local: "authKey"},
			Value:   authKeyValue,
		},
	}
}

func basketAdd(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	basketId, err := getBasketID(r.GetHollywoodId())
	if errors.Is(err, ErrNoBasket) {
		// Create basket
		basketId, err := client.CreateBasket(r.request)
		if err != nil {
			r.ReportError(err)
			return
		}

		_, err = pool.Exec(context.Background(), InsertBasketID, basketId, r.GetHollywoodId())
		if err != nil {
			r.ReportError(err)
		}
		return
	} else if err != nil {
		r.ReportError(err)
		return
	}

	// Edit basket
	if err := client.EditBasket(basketId, r.request); err != nil {
		r.ReportError(err)
	}
}

func basketList(r *Response) {
	basketId, err := getBasketID(r.GetHollywoodId())
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
	basketId, err := getBasketID(r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	// basketNo is the item's position, not a product ID
	basketNo := r.request.URL.Query().Get("basketNo")
	ref, err := client.GetBasketItemIndex(basketId, basketNo)
	if errors.Is(err, justeat.ErrBasketItemIndexNotFound) {
		r.ReportError(justeat.ErrBasketItemNotFound)
		return
	} else if err != nil {
		r.ReportError(err)
		return
	}

	err = client.RemoveItem(basketId, ref)
	if err != nil {
		r.ReportError(err)
	}
}

func basketModify(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	// "Change" always POSTs basket_modify, even if the item's already gone -> handle as a fresh add
	basketId, err := getBasketID(r.GetHollywoodId())
	if errors.Is(err, ErrNoBasket) {
		newBasketId, err := client.CreateBasket(r.request)
		if err != nil {
			r.ReportError(err)
			return
		}

		_, err = pool.Exec(context.Background(), InsertBasketID, newBasketId, r.GetHollywoodId())
		if err != nil {
			r.ReportError(err)
		}
		return
	} else if err != nil {
		r.ReportError(err)
		return
	}

	basketNo := r.request.PostForm.Get("basketNo")
	ref, err := client.GetBasketItemIndex(basketId, basketNo)
	if errors.Is(err, justeat.ErrBasketItemIndexNotFound) {
		if err := client.EditBasket(basketId, r.request); err != nil {
			r.ReportError(err)
		}
		return
	} else if err != nil {
		r.ReportError(err)
		return
	}

	err = client.ModifyBasketItem(basketId, ref, r.request)
	if err != nil {
		r.ReportError(err)
	}
}

func orderDone(r *Response) {
	basketId, err := getBasketID(r.GetHollywoodId())
	if err != nil {
		r.errorCode = http.StatusInternalServerError
		r.ReportError(err)
		return
	}

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.errorCode = http.StatusInternalServerError
		r.ReportError(err)
		return
	}

	err = client.PlaceOrder(r.request, basketId)
	if err != nil {
		PostDiscordWebhook(
			"Performing error failed.",
			fmt.Sprintf("The order was placed by user id %s", r.GetHollywoodId()),
			config.OrderWebhook,
			65311,
		)
		r.errorCode = http.StatusInternalServerError
		r.ReportError(err)
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

	// Post and log successful order!
	PostDiscordWebhook(
		"A successful order has been processed!",
		fmt.Sprintf("The order was placed by user id %s", r.request.Header.Get("X-WiiNo")),
		config.OrderWebhook,
		65311,
	)
}
