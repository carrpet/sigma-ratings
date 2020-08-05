package database

import (
	"database/sql"

	"github.com/lib/pq"
)

type SanctionItem struct {
	LogicalID string `csv:"Entity_LogicalId"`
	WholeName string `csv:"NameAlias_WholeName"`
}

type SanctionMatchResponse struct {
	LogicalID string  `json:"logical_id"`
	WholeName string  `json:"whole_name"`
	Relevance float32 `json:"relevance"`
}

type SanctionResponse struct {
	LogicalID     string   `json:"logicalId"`
	MatchingAlias string   `json:"matchingAlias"`
	OtherAliases  []string `json:"otherAliases"`
	Relevance     float32  `json:"relevance"`
}

func seedSanctionsTxn(sanctions []SanctionItem, db *sql.DB) error {

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

// SeedSanctionsDB initializes the database with the sanctions list
func SeedSanctionsDB(sanctions []SanctionItem, db *sql.DB) error {

	return seedSanctionsTxn(sanctions, db)

	/*
		tableCreateCmd := "CREATE TABLE IF NOT EXISTS sanctions (id SERIAL PRIMARY KEY, logical_id integer NOT NULL, whole_name VARCHAR NOT NULL);"
		_, err := db.Exec(tableCreateCmd)
		if err != nil {
			return err
		}

		// install trgm extension
		trgmExtCmd := "CREATE EXTENSION IF NOT EXISTS pg_trgm"
		_, err = db.Exec(trgmExtCmd)
		if err != nil {
			return err
		}

		//create index to speed up similarity lookup
		indexCreateCmd := "CREATE INDEX IF NOT EXISTS trgm_index ON sanctions USING GIN (whole_name gin_trgm_ops)"
		_, err = db.Exec(indexCreateCmd)
		if err != nil {
			return err
		}

		rows, _ := db.Query("SELECT column_name FROM information_schema.columns WHERE table_name = 'sanctions';")
		log.Println("getting table info")
		for rows.Next() {
			var col_name string
			if err := rows.Scan(&col_name); err != nil {
				log.Fatal(err)
			}
			log.Printf("col_name is: %s", col_name)
		}
		log.Println("end of column info ")

		txn, err := db.Begin()
		if err != nil {
			return err
		}
		stmt, err := txn.Prepare(pq.CopyIn("sanctions", "logical_id", "whole_name"))
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
	*/
}

func QuerySanctionsName(name string, db *sql.DB) ([]SanctionMatchResponse, error) {
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

func GetAliasesForLogicalID(name string, id string, db *sql.DB) ([]SanctionItem, error) {
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
