package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sethvargo/go-envconfig"
	csvlib "github.com/smartystreets/scanners/csv"
)

const configPath = "go/bin/appconfig.yml"

type databaseOpts interface {
	fetchData() ([][]string, error)
}

type dbInfo struct {
	srcURL   string
	host     string
	port     string
	user     string
	dbName   string
	password string
}

func newPGInfo(srcURL, user, dbName, password string) *dbInfo {
	return &dbInfo{srcURL: srcURL, host: "postgres", port: "5432", user: user, dbName: dbName, password: password}
}

type SanctionItem struct {
	LogicalID string `csv:"Entity_LogicalId"`
	WholeName string `csv:"NameAlias_WholeName"`
}

func scanDataToSanctionsList(reader io.Reader) ([]SanctionItem, error) {
	scanner, err := csvlib.NewStructScanner(reader, csvlib.Comma(';'))

	if err != nil {
		return nil, err
	}

	sanctions := []SanctionItem{}
	for scanner.Scan() {
		var sanctionItem SanctionItem
		if err := scanner.Populate(&sanctionItem); err != nil {
			return nil, err
		}
		sanctions = append(sanctions, sanctionItem)
	}

	if err := scanner.Error(); err != nil {
		return nil, err
	}

	return sanctions, nil

}

func (d *dbInfo) fetchData() ([]SanctionItem, error) {
	resp, err := http.Get(d.srcURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return scanDataToSanctionsList(resp.Body)
}

func (d *dbInfo) getDBConnection() (*sql.DB, error) {

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		d.host, d.port, d.user, d.dbName, d.password)

	log.Printf("psqlInfo is: %s", psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	log.Println("got past open!")
	if err != nil {

		log.Printf("error getting db connection, error is: %s", err.Error())
		// if connection succeeds but dbName doesn't exist then create it
		_, err := db.Exec("CREATE DATABASE %s", d.dbName)
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	//not sure how to handle this yet
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected!")

	return db, nil

}

/*
func InsertRecords(items []SanctionItem, db *sql.DB) {

	//
	if err != nil {
		_, err := db.Exec("CREATE DATABASE %s", dbname)
		if err != nil {
			panic(err)
		}
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	// loop through the data and insert into the database
}
*/

func main() {

	var config Config
	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal(err)
	}

	go func() {

		dbInfo := newPGInfo(config.SanctionsBackend.URL, config.Database.User, config.Database.DBName, config.Database.Password)

		_, err := dbInfo.getDBConnection()
		if err != nil {
			log.Printf("Could not get db connection: %s", err.Error())
			return
		}
		_, err = dbInfo.fetchData()
	}()
	// start up the server
	log.Printf("config details: dbName: %s, user: %s", config.Database.DBName, config.Database.User)
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet)
	log.Printf("starting server on localhost:%s", config.FrontEnd.Port)
	var handler http.Handler = r
	log.Fatal(http.ListenAndServe("localhost:"+config.FrontEnd.Port, handler))

}
