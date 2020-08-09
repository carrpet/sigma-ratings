package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
