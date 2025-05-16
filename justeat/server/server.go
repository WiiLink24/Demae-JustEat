package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

const GetOrderParams = `SELECT braintree FROM users WHERE email = $1`

var (
	ctx        = context.Background()
	pool       *pgxpool.Pool
	authConfig *demae.AppAuthConfig
	verifier   *oidc.IDTokenVerifier
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Demae Just Eat Payment Server has encountered an error! Reason: %v\n", err)
	}
}

func RunServer(config *demae.Config, handler http.Handler) {
	// OAuth2 config
	provider, err := oidc.NewProvider(ctx, config.OIDCConfig.Provider)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	authConfig = &demae.AppAuthConfig{
		OAuth2Config: &oauth2.Config{
			ClientID:     config.OIDCConfig.ClientID,
			ClientSecret: config.OIDCConfig.ClientSecret,
			RedirectURL:  config.OIDCConfig.RedirectURL,
			Scopes:       config.OIDCConfig.Scopes,
			Endpoint:     provider.Endpoint(),
		},
		Provider: provider,
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: config.OIDCConfig.ClientID})

	// Open a Postgres pool for this goroutine only.
	dbString := fmt.Sprintf("postgres://%s:%s@%s/%s", config.SQLUser, config.SQLPass, config.SQLAddress, config.SQLDB)
	dbConf, err := pgxpool.ParseConfig(dbString)
	checkError(err)
	pool, err = pgxpool.ConnectConfig(ctx, dbConf)
	checkError(err)

	defer pool.Close()

	// Set up HTTP
	r := gin.Default()
	r.LoadHTMLGlob("./justeat/templates/*")

	r.GET("/login", LoginPage)
	r.GET("/start", StartPanelHandler)
	r.GET("/authorize", FinishPanelHandler)

	auth := r.Group("/")
	auth.Use(AuthenticationMiddleware(verifier))
	{
		auth.GET("/pay", displayPaymentScreen)
		auth.POST("/finalize", finalizePayment)
	}

	fmt.Printf("Starting HTTP connection (%s)...\nJust Eat Payment server connected!\n", config.JustEatAddress)
	log.Fatal(r.Run(config.JustEatAddress))
}

func getBraintreeData(email string) (payload string, hasOrder bool, err error) {
	// Confirm the order exists.
	err = pool.QueryRow(ctx, GetOrderParams, email).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		// TODO: Display no active order for this user.
		return "", false, pgx.ErrNoRows
	} else if err != nil {
		return "", true, err
	}

	hasOrder = true
	return
	// return payload, false, nil
}

func displayPaymentScreen(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		// How did we get here
		c.Status(http.StatusBadRequest)
		return
	}

	payload, hasOrder, err := getBraintreeData(email.(string))
	if !hasOrder && errors.Is(err, pgx.ErrNoRows) {
		// TODO: Display there is no order
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var everything justeat.CombinedBrainTree
	err = json.Unmarshal([]byte(payload), &everything)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "pay.html", gin.H{
		"URL": everything.Head.PaymentResource.RedirectURL,
	})
}

func finalizePayment(c *gin.Context) {
	accEmail, ok := c.Get("email")
	if !ok {
		// How did we get here
		c.Status(http.StatusBadRequest)
		return
	}

	payload, hasOrder, err := getBraintreeData(accEmail.(string))
	if !hasOrder && errors.Is(err, pgx.ErrNoRows) {
		// TODO: Display there is no order
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	var everything justeat.CombinedBrainTree
	err = json.Unmarshal([]byte(payload), &everything)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Pseudo-client
	client := justeat.JEClient{
		Context:      ctx,
		Country:      everything.Country,
		KongAPIURL:   justeat.KongAPIURLs[everything.Country],
		GlobalAPIURL: justeat.GlobalMenuCDNURLs[everything.Country],
		CheckoutURL:  justeat.CheckoutURLs[everything.Country],
		Db:           pool,
	}

	err = client.SetAuth()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	retFirst := justeat.MakePaypalReturnURLFirst(everything.Head)
	nonce, email, payerID, err := client.GetPaypalNonce(everything.BrainTreeConfig, everything.Metadata, everything.BrainTree.AuthorizationFingerprint, retFirst)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	ret := justeat.MakePaypalReturnURL(everything.Head.PaymentResource.PaymentToken, payerID)

	nonce, email, payerID, err = client.GetPaypalNonce(everything.BrainTreeConfig, everything.Metadata, everything.BrainTree.AuthorizationFingerprint, ret)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	err = client.SendPayment(everything.Metadata, nonce, email, payerID, everything.OrderID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}
