// endpoints_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*
Not real tests, just doing some http test framework kata.
See the gorillamux docs for full examples.
*/
func TestEchoHandler(t *testing.T) {
	msg := "42"
	r := strings.NewReader(msg)
	req, err := http.NewRequest("POST", "/echo", r)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(EchoHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if rr.Body.String() != msg {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), msg)
	}
}
