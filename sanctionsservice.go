package main

import (
	"io"
	"net/http"

	csvlib "github.com/smartystreets/scanners/csv"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbName = "sanctions"
)

type databaseOpts interface {
	fetchData() ([][]string, error)
}

type dbInfo struct {
	srcURL string
	host   string
	port   int
	user   string
	dbName string
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
		//fmt.Printf("%#v\n", sanctionItem)
		sanctions = append(sanctions, sanctionItem)
	}

	if err := scanner.Error(); err != nil {
		return nil, err
	}

	return sanctions, nil

}

// returns a file pointer to a csv

func (d *dbInfo) fetchData() ([]SanctionItem, error) {
	resp, err := http.Get(d.srcURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return scanDataToSanctionsList(resp.Body)

}

/*
func ReadIntoDatabase(data [][]string, dbURL string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
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
	dbInfo := &dbInfo{srcURL: "https://sigmaratings.s3.us-east-2.amazonaws.com/eu_sanctions.csv",
		host: host, port: port, user: user, dbName: dbName}

	dbInfo.fetchData()

}
