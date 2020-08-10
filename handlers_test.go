package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carrpet/sigma-ratings/internal/sanction"
)

func TestStatusHandlerReturnsUnavailableUntilItReceivesAMessage(t *testing.T) {
	testRequest := httptest.NewRequest("GET", "/status", nil)
	recorder := httptest.NewRecorder()
	ch := make(chan interface{})
	statusHandler := statusHandlerFactory(ch)
	statusHandler(recorder, testRequest)
	expectedCode := http.StatusServiceUnavailable
	if recorder.Code != expectedCode {
		t.Fatalf("Expected http code %d, received code %d", expectedCode, recorder.Code)
	}

	ch <- struct{}{}
	recorder = httptest.NewRecorder()
	statusHandler(recorder, testRequest)
	expectedCode = http.StatusOK
	if recorder.Code != expectedCode {
		t.Fatalf("Expected http code %d, received code %d", expectedCode, recorder.Code)
	}
}

//Test template for search handler.
func TestSearchHandler(t *testing.T) {
	testRequest := httptest.NewRequest("GET", "/search", nil)
	testRequest.URL.RawQuery = "name=foo"
	mockClient := sanction.NewMockSanctionsClient()
	searchHandler := searchHandlerFactory(mockClient)
	recorder := httptest.NewRecorder()
	searchHandler(recorder, testRequest)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Expected http code %d, received code %d", http.StatusOK, recorder.Code)
	}
}
