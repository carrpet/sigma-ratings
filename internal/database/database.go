package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type SanctionItem struct {
	LogicalID int    `csv:"Entity_LogicalId"`
	WholeName string `csv:"NameAlias_WholeName"`
}

type SanctionMatchResponse struct {
	LogicalID int     `json:"logical_id"`
	WholeName string  `json:"whole_name"`
	Relevance float32 `json:"relevance"`
}

type DBInfo struct {
	host     string
	port     string
	user     string
	dbName   string
	password string
}

type MockDBInfo struct {
	QueryTableExists func() (bool, error)
	InsertTxn        func() error
	QueryName        func() ([]SanctionMatchResponse, error)
	GetAliases       func() ([]SanctionItem, error)
}

type DBOperations interface {
	QuerySanctionsTableExists() (bool, error)
	InsertSanctionsTxn([]SanctionItem) error
	QuerySanctionsName(string) ([]SanctionMatchResponse, error)
	GetAliasesForLogicalID(string, int) ([]SanctionItem, error)
}

var dbInstance *sql.DB

func NewPGInfo(user, dbName, password string) DBOperations {
	return DBInfo{host: "postgres", port: "5432", user: user, dbName: dbName, password: password}
}

func (d MockDBInfo) InsertSanctionsTxn(sanctions []SanctionItem) error {

	return d.InsertTxn()
}

func (d DBInfo) InsertSanctionsTxn(sanctions []SanctionItem) error {

	db, err := d.getInstance()
	if err != nil {
		return err
	}

	txn, err := db.Begin()

	// create main sanctions table
	tableCreateCmd := "CREATE TABLE sanctions (id SERIAL PRIMARY KEY, logical_id integer NOT NULL, whole_name VARCHAR NOT NULL)"
	_, err = txn.Exec(tableCreateCmd)
	if err != nil {
		txn.Rollback()
		return err
	}

	// install trgm extension
	trgmExtCmd := "CREATE EXTENSION pg_trgm"
	_, err = txn.Exec(trgmExtCmd)
	if err != nil {
		txn.Rollback()
		return err
	}

	//create index to speed up similarity lookup
	indexCreateCmd := "CREATE INDEX trgm_index ON sanctions USING GIN (whole_name gin_trgm_ops)"
	_, err = txn.Exec(indexCreateCmd)
	if err != nil {
		txn.Rollback()
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("sanctions", "logical_id", "whole_name"))
	if err != nil {
		txn.Rollback()
		return err
	}

	for _, s := range sanctions {
		_, err = stmt.Exec(s.LogicalID, s.WholeName)
		if err != nil {
			txn.Rollback()
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		txn.Rollback()
		return err
	}

	err = stmt.Close()
	if err != nil {
		txn.Rollback()
		return err
	}

	err = txn.Commit()
	if err != nil {
		txn.Rollback()
		return err
	}
	return nil

}

func (m MockDBInfo) QuerySanctionsName(name string) ([]SanctionMatchResponse, error) {
	return m.QueryName()
}

func (d DBInfo) QuerySanctionsName(name string) ([]SanctionMatchResponse, error) {

	db, err := d.getInstance()
	if err != nil {
		return nil, err
	}
	queryStr := "SELECT logical_id, whole_name, similarity(whole_name, $1) as sml from sanctions WHERE whole_name % $1 ORDER BY sml DESC, whole_name"
	rows, err := db.Query(queryStr, name)
	if err != nil {
		return nil, err
	}

	results := []SanctionMatchResponse{}

	for rows.Next() {
		var resp SanctionMatchResponse
		if err := rows.Scan(&resp.LogicalID, &resp.WholeName, &resp.Relevance); err != nil {
			return nil, err
		}
		results = append(results, resp)
	}
	return results, nil

}

func (m MockDBInfo) GetAliasesForLogicalID(name string, id int) ([]SanctionItem, error) {
	return m.GetAliases()
}

func (d DBInfo) GetAliasesForLogicalID(name string, id int) ([]SanctionItem, error) {

	db, err := d.getInstance()
	if err != nil {
		return nil, err
	}
	queryStr := "SELECT logical_id, whole_name FROM sanctions WHERE logical_id = $1 AND NOT whole_name = '' AND NOT whole_name = $2 ORDER BY whole_name ASC"
	rows, err := db.Query(queryStr, id, name)
	if err != nil {
		return nil, err
	}

	results := []SanctionItem{}
	for rows.Next() {
		var resp SanctionItem
		if err := rows.Scan(&resp.LogicalID, &resp.WholeName); err != nil {
			return nil, err
		}
		results = append(results, resp)
	}
	return results, nil
}

func (d DBInfo) getInstance() (*sql.DB, error) {

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

func (d MockDBInfo) QuerySanctionsTableExists() (bool, error) {

	return d.QueryTableExists()
}

func (d DBInfo) QuerySanctionsTableExists() (bool, error) {

	db, err := d.getInstance()
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
