package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocateHandler(t *testing.T) {
	var locateHandlerTests = []struct {
		Description     string
		InputIP         string
		ExpectedPayload payload
		ExpectedStatus  int
		Handler         http.Handler
	}{
		{
			Description:     "test with working IP located in the wrong country",
			InputIP:         "8.8.8.8",
			ExpectedPayload: payload{Country: "US"},
			ExpectedStatus:  200,
			Handler:         getHandler(apiResponse{Country: "US"}, 200),
		},
		{
			Description:     "test with invalid IP",
			InputIP:         "1.2.3.4",
			ExpectedPayload: payload{Error: "could not locate country info for ip"},
			ExpectedStatus:  404,
			Handler:         getHandler(apiResponse{}, 404),
		},
	}
	server := mockServer()
	defer server.Close()
	client := &http.Client{}
	for _, test := range locateHandlerTests {
		server.Config.Handler = test.Handler
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/locate?ip=%s", test.InputIP), nil)
		w := httptest.NewRecorder()
		locateHandler(client, server.URL).ServeHTTP(w, req)
		if w.Code != test.ExpectedStatus {
			t.Errorf("unexpected status code on %s: %d", test.Description, w.Code)
		}
		var pl payload
		if err := json.NewDecoder(w.Body).Decode(&pl); err != nil {
			t.Error(err)
		}
		if test.ExpectedPayload.Country != "" {
			if pl.Country != test.ExpectedPayload.Country {
				t.Errorf("invalid country, got %s but expected %s", pl.Country, test.ExpectedPayload.Country)
			}
		}
		if test.ExpectedPayload.Error != "" {
			if pl.Error == "" {
				t.Error("error should not be empty")
			}
		}
	}
}

// mockServer contains a default mocked API server
func mockServer() *httptest.Server {
	pl := payload{Country: "DE"}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&pl)
	}))
}

// getHandler gives us a handler with a defined response
func getHandler(ar apiResponse, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(&ar)
	})
}
