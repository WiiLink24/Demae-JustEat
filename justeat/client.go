package justeat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/jackc/pgx/v4/pgxpool"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a generic type that both Just Eat and Skip the Dishes structures must conform to.
type Client interface {
	GetBareRestaurants() ([]demae.CategoryCode, error)
	GetRestaurants(code demae.CategoryCode) ([]demae.BasicShop, error)
	GetRestaurant(id string) (*demae.ShopOne, error)
	GetMenuCategories(id string) ([]demae.Menu, error)
	GetMenuItems(storeId, categoryId string) ([]demae.NestedItem, error)
	GetItemData(shopID, categoryID, itemCode string) ([]demae.ItemOne, float64, error)
	GetMenuGroupID(shopID string) (string, error)
	CreateBasket(r *http.Request) (string, error)
	GetBasket(basketId string, r *http.Request) ([]any, error)
	EditBasket(basketId string, r *http.Request) error
	RemoveItem(basketId string, productId string, r *http.Request) error
	GetAvailableTimes(basketId string) ([]demae.KVFieldWChildren, error)
	PlaceOrder(r *http.Request, basketId string) error
}

type JEClient struct {
	Context      context.Context
	Country      Country
	KongAPIURL   string
	GlobalAPIURL string
	CheckoutURL  string
	Auth         string
	Address      string
	PostalCode   string
	WiiID        string
	Db           *pgxpool.Pool
}

// NewClient constructs either an instance of JEClient or skip.Client.
func NewClient(ctx context.Context, db *pgxpool.Pool, req *http.Request, hollywoodID string) (Client, error) {
	// TODO: Canada detection.
	country, err := GetCountry(req.Header.Get("X-WiiCountryCode"))
	if err != nil {
		return nil, err
	}

	client := &JEClient{
		Context:      ctx,
		Country:      country,
		KongAPIURL:   KongAPIURLs[country],
		GlobalAPIURL: GlobalMenuCDNURLs[country],
		CheckoutURL:  CheckoutURLs[country],
		Address:      req.Header.Get("X-Address"),
		PostalCode:   req.Header.Get("X-PostalCode"),
		WiiID:        hollywoodID,
		Db:           db,
	}

	err = client.SetAuth()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (j *JEClient) SetAuth() error {
	// First we want to see if our auth key has expired.
	var expiresAt time.Time
	var refreshToken string
	var acr string
	row := j.Db.QueryRow(j.Context, QueryAuthExpiryTime, j.WiiID)
	err := row.Scan(&expiresAt, &refreshToken, &acr)
	if err != nil {
		return err
	}

	var auth string
	if expiresAt.Before(time.Now().UTC()) {
		// Generate the new auth token
		auth, err = j.refreshAuthToken(refreshToken, acr, j.WiiID)
		if err != nil {
			return err
		}
	} else {
		row = j.Db.QueryRow(j.Context, QueryUserAuth, j.WiiID)
		err = row.Scan(&auth)
		if err != nil {
			return err
		}
	}

	j.Auth = auth
	return nil
}

func (j *JEClient) refreshAuthToken(refreshToken, acr, hash string) (string, error) {
	_url := fmt.Sprintf("%s/identity/connect/token", j.KongAPIURL)

	payload := url.Values{}
	payload.Set("refresh_token", refreshToken)
	payload.Set("grant_type", "refresh_token")
	payload.Set("scope", "openid mobile_scope offline_access")
	payload.Set("acr_values", acr)

	resp, err := j.unauthorizedPost(_url, payload)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var temp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	err = json.Unmarshal(body, &temp)
	if err != nil {
		return "", err
	}

	// Now save to database
	auth := fmt.Sprintf("Bearer %s", temp.AccessToken)
	expiresAt := time.Now().UTC().Add(time.Second * time.Duration(temp.ExpiresIn))

	_, err = j.Db.Exec(j.Context, UpdateAuthToken, auth, temp.RefreshToken, expiresAt, hash)
	if err != nil {
		return "", err
	}

	return auth, nil
}

func GetCountry(countryCode string) (Country, error) {
	switch countryCode {
	// TODO: Italy is actually 089 but i am testing my code
	case "110":
		return UnitedKingdom, nil
	default:
		return UnitedKingdom, nil
	}
}
