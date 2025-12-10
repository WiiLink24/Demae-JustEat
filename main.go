package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat/server"
	"github.com/WiiLink24/DemaeJustEat/logger"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/logrusorgru/aurora/v4"
	"github.com/redis/go-redis/v9"
	"github.com/remizovm/geonames"
	"github.com/remizovm/geonames/models"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	pool          *pgxpool.Pool
	ctx           = context.Background()
	config        *demae.Config
	rdb           *redis.Client
	geonameCities map[int]*models.Feature
	geonameStates []*models.AdminCode
	sentryHandler *sentryhttp.Handler
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Demae Just Eat server has encountered an error! Reason: %v\n", err)
	}
}

func main() {
	// Load the config
	rawConfig, err := os.ReadFile("./config.xml")
	checkError(err)

	config = &demae.Config{}
	err = xml.Unmarshal(rawConfig, config)
	checkError(err)

	// Before we do anything, init Sentry to capture all errors.
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              config.SentryDSN,
		Debug:            config.IsDebug,
		EnableTracing:    false,
		TracesSampleRate: 1.0,
	})
	checkError(err)
	defer sentry.Flush(2 * time.Second)

	sentryHandler = sentryhttp.New(sentryhttp.Options{})

	logger.SetDebug(config.IsDebug)

	// Initialize database
	dbString := fmt.Sprintf("postgres://%s:%s@%s/%s", config.SQLUser, config.SQLPass, config.SQLAddress, config.SQLDB)
	dbConf, err := pgxpool.ParseConfig(dbString)
	checkError(err)
	pool, err = pgxpool.ConnectConfig(ctx, dbConf)
	checkError(err)

	// Ensure this Postgresql connection is valid.
	defer pool.Close()

	// Initialize Geonames.
	client := geonames.Client{}
	geonameCities, err = client.Cities15000()
	checkError(err)

	geonameStates, err = client.Admin1CodesASCII()
	checkError(err)

	rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword,
		DB:       0,
	})

	r := NewRoute()
	nwapi := r.HandleGroup("nwapi.php")
	{
		nwapi.NormalResponse("webApi_document_template", documentTemplate)
		nwapi.NormalResponse("webApi_area_list", areaList)
		nwapi.MultipleRootNodes("webApi_category_list", categoryList)
		nwapi.NormalResponse("webApi_area_shopinfo", shopInfo)
		nwapi.NormalResponse("webApi_shop_list", shopList)
		nwapi.MultipleRootNodes("webApi_shop_one", shopOne)
		nwapi.MultipleRootNodes("webApi_menu_list", menuList)
		nwapi.MultipleRootNodes("webApi_item_list", itemList)
		nwapi.MultipleRootNodes("webApi_item_one", itemOne)
		nwapi.MultipleRootNodes("webApi_Authkey", authKey)
		nwapi.MultipleRootNodes("webApi_basket_add", basketAdd)
		nwapi.MultipleRootNodes("webApi_basket_list", basketList)
		nwapi.MultipleRootNodes("webApi_basket_delete", basketDelete)
		nwapi.MultipleRootNodes("webApi_basket_reset", basketReset)
		nwapi.MultipleRootNodes("webApi_validate_condition", func(r *Response) {})
		nwapi.NormalResponse("webApi_order_done", orderDone)
		nwapi.NormalResponse("webApi_inquiry_done", inquiryDone)
	}

	logo := r.HandleGroup("logoimg2")
	{
		logo.ServeImage(func(r *Response) {
			// Remove "l_" from the URL.
			path := strings.Replace(r.request.URL.Path, "l_", "", 1)
			paths := strings.Split(path, "/")

			data, err := os.ReadFile(fmt.Sprintf("logos/%s", paths[2]))
			if err != nil {
				log.Println("failed to read image")
			}

			(*r.writer).Write(data)
		})
	}

	itemImg := r.HandleGroup("itemimg")
	{
		itemImg.ServeImage(func(r *Response) {
			path := strings.Replace(r.request.URL.Path, "l_", "", 1)
			splitUrl := strings.Split(path, "/")

			img, err := os.ReadFile(fmt.Sprintf("logos/%s/%s", splitUrl[2], splitUrl[3]))
			if err != nil {
				log.Println("failed to read image")
			}

			(*r.writer).Write(img)
			return
		})
	}

	// Start the Demae Channel server as well as the Just Eat payment server.
	go server.RunServer(config)

	fmt.Printf("Starting HTTP connection (%s)...\n%s\n", aurora.Yellow(config.DemaeAddress), aurora.Green("Demae Just Eat server connected!"))
	log.Fatal(http.ListenAndServe(config.DemaeAddress, r.Handle()))
}
