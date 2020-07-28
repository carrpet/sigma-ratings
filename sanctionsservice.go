package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	csvlib "github.com/smartystreets/scanners/csv"
	"gopkg.in/yaml.v2"
)

const configPath = "go/bin/appconfig.yml"

type databaseOpts interface {
	fetchData() ([][]string, error)
}

type dbInfo struct {
	srcURL string
	host   string
	port   string
	user   string
	dbName string
}

func newPGInfo(srcURL, user, dbName string) *dbInfo {
	return &dbInfo{srcURL: srcURL, host: "postgres", port: "5432", user: user, dbName: dbName}
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

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		d.host, d.port, d.user, d.dbName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {

		fmt.Printf("error getting db connection, error is: %s", err.Error())
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

	fmt.Println("Successfully connected!")

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

func readConfig(cfg *Config) error {
	//f, err := os.Open(configPath)
	conf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	conf = []byte(os.ExpandEnv(string(conf)))
	//err = yaml.NewDecoder(f).Decode(cfg)
	//if err != nil {
	//	return err
	//}
	//return nil
	if err := yaml.Unmarshal(conf, cfg); err != nil {
		return err
	}
	return nil
}

func main() {

	var config Config
	err := readConfig(&config)
	if err != nil {
		log.Fatal(err.Error())
	}

	go func() {

		dbInfo := newPGInfo(config.SanctionsBackend.URL, config.Database.User, config.Database.DBName)

		_, err = dbInfo.getDBConnection()
		if err != nil {
			fmt.Println("Could not get db connection")
		}
		_, err = dbInfo.fetchData()
	}()

	// start up the server
	fmt.Printf("config details: dbName: %s, user: %s", config.Database.DBName, config.Database.User)
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet)
	fmt.Println("starting server on something!")
	var handler http.Handler = r
	log.Fatal(http.ListenAndServe("localhost:"+"8080", handler))

}
