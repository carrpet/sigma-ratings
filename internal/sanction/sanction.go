package sanction

import (
	"io"
	"log"
	"net/http"

	"github.com/carrpet/sigma-ratings/internal/database"
	"github.com/smartystreets/scanners/csv"
)

// SanctionsDB defines sanctions operations.
type SanctionsDB interface {
	InitSanctionsData() error
	GetRelevantSanctionAndAliases(string) ([]Response, error)
}

// SanctionsBackend defines operations on sanctions source.
type SanctionsBackend interface {
	GetSanctionsList() ([]database.SanctionItem, error)
}

// Response represents sanctions api response.
type Response struct {
	LogicalID     int      `json:"logicalId"`
	MatchingAlias string   `json:"matchingAlias"`
	OtherAliases  []string `json:"otherAliases"`
	Relevance     float32  `json:"relevance"`
}

// SanctionsClient represents client dependencies.
type SanctionsClient struct {
	DBInfo       database.DBOperations
	SanctionsURL SanctionsBackend
}

// SanctionsURL implements SanctionsBackend interface
type SanctionsURL struct {
	URL string
}

// NewSanctionsClient is a client constructor.
func NewSanctionsClient(dbName, user, password, url string) SanctionsDB {
	dbInfo := database.NewPGInfo(user, dbName, password)
	return SanctionsClient{DBInfo: dbInfo, SanctionsURL: SanctionsURL{URL: url}}
}

// InitSanctionsData retrieves sanctions from the data source and inserts it into
// the application database.
func (c SanctionsClient) InitSanctionsData() error {
	exists, _ := c.DBInfo.QuerySanctionsTableExists()
	if !exists {
		log.Println("sanctions table doesn't exist, seeding db")
		items, err := c.SanctionsURL.GetSanctionsList()
		if err != nil {
			return err
		}

		err = c.DBInfo.InsertSanctionsTxn(items)
		if err != nil {
			return err
		}
	}

	return nil

}

// GetSanctionsList retrieves sanctions from the data source.
func (s SanctionsURL) GetSanctionsList() ([]database.SanctionItem, error) {

	resp, err := http.Get(s.URL)
	log.Printf("Retrieving sanctions list from: %s", s.URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return scanDataToSanctionsList(resp.Body)

}

// GetRelevantSanctionAndAliases searches for the closest match to provided name
// and returns the matching sanction, relevance, and the other aliases for that entry.
func (c SanctionsClient) GetRelevantSanctionAndAliases(name string) ([]Response, error) {

	sanctions, err := c.DBInfo.QuerySanctionsName(name)
	if err != nil {
		return nil, err
	}

	results := []database.SanctionMatchResponse{}
	for _, item := range sanctions {
		if item.Relevance < 1 && len(results) == 0 {
			results = append(results, item)
			break
		} else if item.Relevance < 1 {
			break
		} else {
			results = append(results, item)
		}
	}

	resp := []Response{}
	for _, r := range results {
		aliasList := []string{}
		aliases, err := c.DBInfo.GetAliasesForLogicalID(r.WholeName, r.LogicalID)
		if err != nil {
			return nil, err
		}
		for _, s := range aliases {
			aliasList = append(aliasList, s.WholeName)
		}
		resp = append(resp, Response{LogicalID: r.LogicalID, MatchingAlias: r.WholeName, OtherAliases: aliasList, Relevance: r.Relevance})
	}
	return resp, nil
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

/* Code to support mocking for unit tests */

// MockSanctionsBackend is a mock.
type MockSanctionsBackend struct {
}

// GetSanctionsList is a mock function.
func (m MockSanctionsBackend) GetSanctionsList() ([]database.SanctionItem, error) {
	return []database.SanctionItem{}, nil
}

// NewMockSanctionsClient returns a mock client.
func NewMockSanctionsClient() SanctionsDB {
	queryNameFunc := func() ([]database.SanctionMatchResponse, error) {
		return []database.SanctionMatchResponse{}, nil
	}
	getAliasesFunc := func() ([]database.SanctionItem, error) {
		return []database.SanctionItem{}, nil
	}
	return SanctionsClient{DBInfo: database.MockDBInfo{QueryName: queryNameFunc, GetAliases: getAliasesFunc}, SanctionsURL: MockSanctionsBackend{}}
}
