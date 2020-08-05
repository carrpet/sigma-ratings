package sanction

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/carrpet/sigma-ratings/internal/database"
	"github.com/smartystreets/scanners/csv"
)

var dbInstance *sql.DB

type SanctionsOpts interface {
	PopulateSanctions([]database.SanctionItem) error
}

type SanctionsBackendOpts interface {
	GetSanctionsList(url string) ([]database.SanctionItem, error)
}

type DBInfo struct {
	host     string
	port     string
	user     string
	dbName   string
	password string
}

func NewPGInfo(user, dbName, password string) *DBInfo {
	return &DBInfo{host: "postgres", port: "5432", user: user, dbName: dbName, password: password}
}

func (d *DBInfo) QuerySanctionsTableExistence() (bool, error) {

	db, err := d.getDBConnection()
	if err != nil {
		return false, err
	}

	tableExistsQuery := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sanctions')"
	row := db.QueryRow(tableExistsQuery)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("query sanction table returned error: %s", err.Error())
		return false, err

	}
	return exists, nil

}

func (d *DBInfo) PopulateSanctions(items []database.SanctionItem) error {

	//establish db connection
	db, err := d.getDBConnection()
	if err != nil {
		return err
	}
	err = database.SeedSanctionsDB(items, db)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBInfo) GetSanctionsList(url string) ([]database.SanctionItem, error) {

	resp, err := http.Get(url)
	log.Printf("Retrieving sanctions list from: %s", url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return scanDataToSanctionsList(resp.Body)

}
func scanDataToSanctionsList(reader io.Reader) ([]database.SanctionItem, error) {
	scanner, err := csv.NewStructScanner(reader, csv.Comma(';'))

	if err != nil {
		return nil, err
	}

	sanctions := []database.SanctionItem{}
	for scanner.Scan() {
		var sanctionItem database.SanctionItem
		if err := scanner.Populate(&sanctionItem); err != nil {
			return nil, err
		}
		sanctions = append(sanctions, sanctionItem)
	}

	if err := scanner.Error(); err != nil {
		return nil, err
	}

	log.Println("Retrieved sanctions from sanctions backend")

	return sanctions, nil

}

func (d *DBInfo) getDBConnection() (*sql.DB, error) {

	if dbInstance == nil {
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
			d.host, d.port, d.user, d.dbName, d.password)

		log.Printf("psqlInfo is: %s", psqlInfo)
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			return nil, err
		}

		err = db.Ping()
		if err != nil {
			return nil, err
		}
		dbInstance = db
	}

	return dbInstance, nil

}

func (d *DBInfo) GetRelevantSanctionAndAliases(name string) ([]database.SanctionResponse, error) {
	sanctions, err := d.QuerySanctionsByName(name)
	if err != nil {
		return nil, err
	}
	db, err := d.getDBConnection()
	if err != nil {
		return nil, err
	}

	results := []database.SanctionMatchResponse{}
	for _, item := range sanctions {
		results = append(results, item)
		if item.Relevance < 1 {
			break
		}
	}

	resp := []database.SanctionResponse{}
	for _, r := range results {
		aliasList := []string{}
		aliases, err := database.GetAliasesForLogicalID(r.WholeName, r.LogicalID, db)
		if err != nil {
			return nil, err
		}
		for _, s := range aliases {
			aliasList = append(aliasList, s.WholeName)
		}
		resp = append(resp, database.SanctionResponse{LogicalID: r.LogicalID, MatchingAlias: r.WholeName, OtherAliases: aliasList, Relevance: r.Relevance})
	}
	return resp, nil
}

func (d *DBInfo) QuerySanctionsByName(name string) ([]database.SanctionMatchResponse, error) {
	db, err := d.getDBConnection()
	if err != nil {
		return nil, err
	}
	return database.QuerySanctionsName(name, db)
}
