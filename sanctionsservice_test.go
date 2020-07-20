package main

import (
	"strings"
	"testing"
)

type dbMock struct {
}

/*
func (d *dbMock) fetchData() ([]string, error) {
	return nil, nil
}
*/

//TestFetchSanctions goes out to the database, reads csv containing sanctions into memory and
// inserts it into the database
func TestScanDataToSanctionsList(t *testing.T) {
	testdata := []string{`Entity_LogicalId;NameAlias_WholeName`, `143; Saddam H`, `159; John Q`}
	testReader := strings.NewReader(strings.Join(testdata, "\n"))
	result, err := scanDataToSanctionsList(testReader)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for i, s := range result {
		fields := strings.Split(testdata[i+1], ";")
		if s.LogicalID != fields[0] {
			t.Fatalf("failed logicalId to sanctionItem conversion, expected: %s, actual: %s", fields[0], s.LogicalID)
		}
		if s.WholeName != fields[1] {
			t.Fatalf("failed WholeName to sanctionItem conversion, expected: %s, actual: %s", fields[1], s.WholeName)
		}
	}
}

//Integration Tests
func TestFetchData(t *testing.T) {

	d := &dbInfo{srcURL: "https://sigmaratings.s3.us-east-2.amazonaws.com/eu_sanctions.csv"}
	data, err := d.fetchData()
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, d := range data {
		t.Logf("%#v\n", d)
	}
}
