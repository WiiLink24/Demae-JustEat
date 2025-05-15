package justeat

import (
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"log"
	"net/http"
)

func ServerMain(config *demae.Config, handler http.Handler) {
	fmt.Printf("Starting HTTP connection (%s)...\nJust Eat Payment server connected!\n", config.JustEatAddress)
	log.Fatal(http.ListenAndServe(config.JustEatAddress, handler))
}
