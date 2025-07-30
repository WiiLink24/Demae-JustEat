package server

import (
	"context"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/logrusorgru/aurora/v4"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

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
	if config.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.LoadHTMLGlob("./justeat/templates/*")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/login")
	})

	r.GET("/login", LoginPage)
	r.GET("/start", StartPanelHandler)
	r.GET("/authorize", FinishPanelHandler)
	r.GET("/userdatalogin.json", getLoginData)
	r.GET("/2fadata.json", get2FAData)

	auth := r.Group("/")
	auth.Use(AuthenticationMiddleware(verifier))
	{
		auth.GET("/pay", displayPaymentScreen)
		auth.POST("/finalize", finalizePayment)
		auth.POST("/cancel", cancelPayment)
	}

	authLink := r.Group("/")
	authLink.Use(AuthenticationLinkerMiddleware(verifier))
	{
		authLink.POST("/link", saveUserData)
	}

	fmt.Printf("Starting HTTP connection (%s)...\n%s\n", aurora.Yellow(config.JustEatAddress), aurora.Green("Just Eat Payment server connected!"))
	log.Fatal(r.Run(config.JustEatAddress))
}
