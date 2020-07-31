package database

import (
	"database/sql"

	"github.com/lib/pq"
)

type SanctionItem struct {
	LogicalID string `csv:"Entity_LogicalId"`
	WholeName string `csv:"NameAlias_WholeName"`
}

// SeedSanctionsDB initializes the database with the sanctions list
func SeedSanctionsDB(sanctions []SanctionItem, db *sql.DB) error {
	txn, err := db.Begin()

	// create main sanctions table
	tableCreateCmd := "CREATE TABLE sanctions (id SERIAL PRIMARY KEY, logicalID integer NOT NULL, wholeName VARCHAR NOT NULL)"
	_, err = txn.Exec(tableCreateCmd)
	if err != nil {
		return err
	}

	stmt, err := txn.Prepare(pq.CopyIn("sanctions", "logicalID", "wholeName"))
	if err != nil {
		return err
	}

	for _, s := range sanctions {
		_, err = stmt.Exec(s.LogicalID, s.WholeName)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil

}

func InsertRecords(items []SanctionItem, db *sql.DB) error {

	// loop through the data and insert into the database
	err := SeedSanctionsDB(items, db)
	if err != nil {
		return err
	}
	return nil
}
