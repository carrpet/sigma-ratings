package main

import (
	"net/http/httptest"
	"testing"

	"github.com/carrpet/sigma-ratings/internal/sanction"
)

func TestStatusHandler(t *testing.T) {
	testRequest := httptest.NewRequest("GET", "/search", nil)
	testRequest.URL.RawQuery = "name=foobar"
	recorder := httptest.NewRecorder()
	mockDB := sanction.MockDBInfo{}
	sHandler := searchHandlerFactory(mockDB)
	sHandler(recorder, testRequest)
	if recorder.Code != 200 {
		t.Fatal()
	}
}
