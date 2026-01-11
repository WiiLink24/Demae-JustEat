package justeat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
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
	Auth         string
	Address      string
	PostalCode   string
	WiiID        string
	DeviceModel  string
	Db           *pgxpool.Pool
	rdb          *redis.Client
}

// NewClient constructs either an instance of JEClient or skip.Client.
func NewClient(ctx context.Context, db *pgxpool.Pool, req *http.Request, hollywoodID string, rdb *redis.Client) (Client, error) {
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
		Address:      req.Header.Get("X-Address"),
		PostalCode:   req.Header.Get("X-PostalCode"),
		WiiID:        hollywoodID,
		Db:           db,
		rdb:          rdb,
	}

	err = client.SetAuth()
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, NotLinked
	} else if err != nil {
		return nil, err
	}

	return client, nil
}

func (j *JEClient) SetAuth() error {
	// First we want to see if our auth key has expired.
	var expiresAt time.Time
	var refreshToken string
	var acr string
	row := j.Db.QueryRow(j.Context, QueryUserData, j.WiiID)
	err := row.Scan(&j.Auth, &expiresAt, &refreshToken, &acr, &j.DeviceModel)
	if err != nil {
		return err
	}

	if expiresAt.Before(time.Now().UTC()) {
		// Generate the new auth token
		j.Auth, err = j.refreshAuthToken(refreshToken, acr, j.WiiID)
		if err != nil {
			return err
		}
	}

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
	case "066":
		return Austria, nil
	case "078":
		return Germany, nil
	case "082":
		return Ireland, nil
	case "083":
		return Italy, nil
	case "105":
		return Spain, nil
	case "110":
		return UnitedKingdom, nil
	default:
		return UnitedKingdom, nil
	}
}
