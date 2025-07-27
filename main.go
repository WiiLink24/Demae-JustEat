package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat/server"
	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/remizovm/geonames"
	"github.com/remizovm/geonames/models"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	pool          *pgxpool.Pool
	ctx           = context.Background()
	config        *demae.Config
	geonameCities map[int]*models.Feature
	geonameStates []*models.AdminCode
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
		Debug:            true,
		TracesSampleRate: 1.0,
	})
	checkError(err)
	defer sentry.Flush(2 * time.Second)

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

	fmt.Printf("Starting HTTP connection (%s)...\nNot using the usual port for HTTP?\nBe sure to use a proxy, otherwise the Wii can't connect!\n", config.DemaeAddress)
	r := NewRoute()
	nwapi := r.HandleGroup("nwapi.php")
	{
		nwapi.NormalResponse("webApi_document_template", documentTemplate)
		nwapi.NormalResponse("webApi_area_list", areaList)
		nwapi.MultipleRootNodes("webApi_category_list", categoryList)
		nwapi.NormalResponse("webApi_area_shopinfo", func(r *Response) {})
		nwapi.NormalResponse("webApi_shop_list", shopList)
		nwapi.MultipleRootNodes("webApi_shop_one", shopOne)
		nwapi.MultipleRootNodes("webApi_menu_list", menuList)
		nwapi.MultipleRootNodes("webApi_item_list", itemList)
		nwapi.MultipleRootNodes("webApi_item_one", itemOne)
		nwapi.MultipleRootNodes("webApi_Authkey", authKey)
		nwapi.MultipleRootNodes("webApi_basket_add", basketAdd)
		nwapi.MultipleRootNodes("webApi_basket_list", basketList)
		nwapi.MultipleRootNodes("webApi_basket_delete", basketDelete)
		nwapi.MultipleRootNodes("webApi_validate_condition", func(r *Response) {})
		nwapi.NormalResponse("webApi_order_done", orderDone)
		nwapi.NormalResponse("webApi_inquiry_done", inquiryDone)
	}

	logo := r.HandleGroup("logoimg2")
	{
		logo.ServeImage(func(r *Response) {
			paths := strings.Split(r.request.URL.Path, "/")

			data, err := os.ReadFile(fmt.Sprintf("logos/%s", paths[2]))
			if err != nil {
				log.Println("failed to read image")
			}

			(*r.writer).Write(data)
		})
	}

	// Start the Demae Channel server as well as the Just Eat payment server.
	actions := []func(*demae.Config, http.Handler){demaeMain, server.RunServer}
	handlers := []http.Handler{r.Handle(), nil}
	wg := &sync.WaitGroup{}
	wg.Add(len(actions))
	for i, action := range actions {
		go func(a func(*demae.Config, http.Handler)) {
			defer wg.Done()
			a(config, handlers[i])
		}(action)
	}

	wg.Wait()
}

func demaeMain(config *demae.Config, handler http.Handler) {
	log.Fatal(http.ListenAndServe(config.DemaeAddress, handler))
}
