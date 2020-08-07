package main

import (
	"log"
	"net/http"

	"encoding/json"

	"github.com/carrpet/sigma-ratings/internal/sanction"
)

func statusHandlerFactory(availableCh chan interface{}) func(w http.ResponseWriter, r *http.Request) {

	currentStatus := http.StatusServiceUnavailable
	go func() {
		<-availableCh
		currentStatus = http.StatusOK
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(currentStatus)

	}
}

func searchHandlerFactory(client sanction.SanctionsDB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println("Could not parse request form")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		nameParam := r.Form.Get("name")

		results, err := client.GetRelevantSanctionAndAliases(nameParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resultBytes, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resultBytes)
	}

}
