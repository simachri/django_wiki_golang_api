package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joho/godotenv"
)

func TestApiRoutes(t *testing.T) {
	cases := []struct {
		descr     string
		httpType  string
		endpoint  string
		expCode   int
		expString string
	}{
		{"GET endpoint /ping", "GET", "/ping", 200, "pong"},
		{"GET endpoint /db/health", "GET", "/db/heatlh", 200, ""},
	}

	// A DB connection requires environment variables.
	godotenv.Load("../../.env")
	router := setupRouter()
	for _, tc := range cases {
		t.Run(tc.descr, func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.expCode {
				t.Fatalf("expected return code %v, but got %v", tc.expCode, w.Code)
			}
			if tc.expString != "" && w.Body.String() != tc.expString {
				t.Fatalf("expected return string %v, but got %v", tc.expString, w.Body.String())
			}
		})
	}
}
