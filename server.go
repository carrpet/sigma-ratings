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

	availableCh := make(chan interface{})
	sClient := sanction.NewSanctionsClient(config.Database.DBName, config.Database.User, config.Database.Password, config.SanctionsBackend.URL)

	go func() {

		err := sClient.InitSanctionsData()
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		availableCh <- struct{}{}

	}()
	// start up the server
	log.Printf("config details: dbName: %s, user: %s", config.Database.DBName, config.Database.User)
	r := mux.NewRouter()
	r.HandleFunc("/status", statusHandlerFactory(availableCh)).Methods(http.MethodGet)
	r.HandleFunc("/search", searchHandlerFactory(sClient)).Methods(http.MethodGet)
	log.Printf("starting server on 0.0.0.0:%s", config.FrontEnd.Port)
	var handler http.Handler = r
	log.Fatal(http.ListenAndServe("0.0.0.0:"+config.FrontEnd.Port, handler))

}
