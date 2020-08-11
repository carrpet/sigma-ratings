package sanction

import (
	"testing"

	"github.com/carrpet/sigma-ratings/internal/database"
)

func mockSanctionsClient() SanctionsClient {
	return SanctionsClient{}
}

func TestGetRelevantSanctionsAndAliasesReturnsRelevance1(t *testing.T) {

	queryResp := []database.SanctionMatchResponse{
		{LogicalID: 32, Relevance: 1},
		{LogicalID: 34, Relevance: 1},
		{LogicalID: 80, Relevance: 0.6}}
	querySanctionsResp := func() ([]database.SanctionMatchResponse, error) {
		return queryResp, nil
	}
	getAliasesResp := func() ([]database.SanctionItem, error) {
		return []database.SanctionItem{}, nil
	}

	mockDB := database.MockDBInfo{QueryName: querySanctionsResp, GetAliases: getAliasesResp}
	client := SanctionsClient{DBInfo: mockDB, SanctionsURL: SanctionsURL{URL: "foo.com"}}
	result, err := client.GetRelevantSanctionAndAliases("foo")
	if err != nil {
		t.Fatal("Expected success")
	}
	if len(result) != 2 {
		t.Fatalf("Expected result 2, received result length: %d", len(result))
	}
	for i := 0; i < 2; i++ {
		if queryResp[i].LogicalID != result[i].LogicalID {
			t.Fatalf("Expected logicalID %d, received %d", queryResp[i].LogicalID, result[i].LogicalID)
		}
	}
}

func TestGetRelevantSanctionsAndAliasesFindsOtherAliases(t *testing.T) {

	queryNameResp := func() ([]database.SanctionMatchResponse, error) {
		return []database.SanctionMatchResponse{
			{LogicalID: 32, Relevance: 0.6, WholeName: "b"},
			{LogicalID: 34, Relevance: 0.3},
			{LogicalID: 80, Relevance: 0.3}}, nil
	}

	getAliasesResp := func() ([]database.SanctionItem, error) {
		return []database.SanctionItem{
			{LogicalID: "32", WholeName: "a"},
			{LogicalID: "32", WholeName: "c"},
			{LogicalID: "32", WholeName: "d"}}, nil
	}

	mockDB := database.MockDBInfo{QueryName: queryNameResp, GetAliases: getAliasesResp}
	client := SanctionsClient{DBInfo: mockDB, SanctionsURL: SanctionsURL{URL: "foo.com"}}
	result, err := client.GetRelevantSanctionAndAliases("foo")
	if err != nil {
		t.Fatal("Expected success")
	}
	if len(result) != 1 {
		t.Fatalf("Expected result 1, received result length: %d", len(result))
	}

	otherAliases := result[0].OtherAliases
	if len(otherAliases) != 3 {
		t.Fatalf("Expected length otherAliases to be 3, received %d", len(otherAliases))
	}

}
