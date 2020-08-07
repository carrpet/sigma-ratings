package sanction

import (
	"io"
	"log"
	"net/http"

	"github.com/carrpet/sigma-ratings/internal/database"
	"github.com/smartystreets/scanners/csv"
)

type SanctionsDB interface {
	InitSanctionsData() error
	GetRelevantSanctionAndAliases(string) ([]SanctionResponse, error)
}

type SanctionResponse struct {
	LogicalID     int      `json:"logicalId"`
	MatchingAlias string   `json:"matchingAlias"`
	OtherAliases  []string `json:"otherAliases"`
	Relevance     float32  `json:"relevance"`
}

type SanctionsClient struct {
	DBInfo       database.DBOperations
	SanctionsURL SanctionsBackend
}

type SanctionsBackend interface {
	GetSanctionsList() ([]database.SanctionItem, error)
}

type SanctionsURL struct {
	URL string
}

func NewSanctionsClient(dbName, user, password, url string) SanctionsDB {
	dbInfo := database.NewPGInfo(user, dbName, password)
	return SanctionsClient{DBInfo: dbInfo, SanctionsURL: SanctionsURL{URL: url}}

}

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

func (s SanctionsURL) GetSanctionsList() ([]database.SanctionItem, error) {

	resp, err := http.Get(s.URL)
	log.Printf("Retrieving sanctions list from: %s", s.URL)
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

func (c SanctionsClient) GetRelevantSanctionAndAliases(name string) ([]SanctionResponse, error) {

	sanctions, err := c.DBInfo.QuerySanctionsName(name)
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

	resp := []SanctionResponse{}
	for _, r := range results {
		aliasList := []string{}
		aliases, err := c.DBInfo.GetAliasesForLogicalID(r.WholeName, r.LogicalID)
		if err != nil {
			return nil, err
		}
		for _, s := range aliases {
			aliasList = append(aliasList, s.WholeName)
		}
		resp = append(resp, SanctionResponse{LogicalID: r.LogicalID, MatchingAlias: r.WholeName, OtherAliases: aliasList, Relevance: r.Relevance})
	}
	return resp, nil
}
