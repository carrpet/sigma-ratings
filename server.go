package main

import (
	"context"
	"log"
	"net/http"

	"github.com/carrpet/sigma-ratings/internal/sanction"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sethvargo/go-envconfig"
)

const configPath = "go/bin/appconfig.yml"

func main() {

	var config Config
	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal(err)
	}

	go func() {
		pgInfo := sanction.NewPGInfo(config.Database.User, config.Database.DBName, config.Database.Password)
		items, err := pgInfo.GetSanctionsList(config.SanctionsBackend.URL)
		if err != nil {
			log.Printf("Could not retrieve sanctions from source %s: %s", config.SanctionsBackend.URL, err.Error())
			return
		}

		err = pgInfo.PopulateSanctions(items)
		if err != nil {
			log.Printf("Could not population sanctions DB: %s", err.Error())
			return
		}

	}()
	// start up the server
	log.Printf("config details: dbName: %s, user: %s", config.Database.DBName, config.Database.User)
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet)
	log.Printf("starting server on localhost:%s", config.FrontEnd.Port)
	var handler http.Handler = r
	log.Fatal(http.ListenAndServe("localhost:"+config.FrontEnd.Port, handler))

}
