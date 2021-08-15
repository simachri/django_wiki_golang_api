package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	m "coco-life.de/wapi/internal"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestApiRoutes(t *testing.T) {
	cases := []struct {
		descr       string
		httpType    string
		endpoint    string
		expCode     int
		expString   string
		expResponse m.Response
	}{
		{"GET endpoint /ping", "GET", "/ping", 200, "pong", nil},
		{"GET endpoint /db/health", "GET", "/db/health", 200, "", nil},
		{"GET endpoint /articles/1", "GET", "/articles/1", 200, "",
			&m.Article{
				ID:      1,
				Title:   "Hello, hello, hello",
				Content: "# Hello World"}},
	}

	// A DB connection requires environment variables.
	godotenv.Load("../../.env")
	router := setupRouter()
	for _, tc := range cases {
		t.Run(tc.descr, func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.endpoint, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expCode, w.Code, "expected return code %v, but got %v", tc.expCode, w.Code)
			if tc.expString != "" {
				assert.Equal(t, tc.expString, w.Body.String(), "expected return string %v, but got %v", tc.expString, w.Body.String())
			}
			if tc.expResponse != nil {
				fmt.Printf("Response body: %v\n", w.Body.String())
				respType := reflect.TypeOf(tc.expResponse).Elem()
				data := reflect.New(respType).Interface()
				err := json.Unmarshal([]byte(w.Body.String()), &data)
				assert.Nil(t, err)
				obj := data.(m.Response)
                assert.True(t, tc.expResponse.Equals(obj),
                    fmt.Sprintf(
                            "JSON response differs.\n" +
                            "Exp: %v\n" +
                            "Act: %v\n", tc.expResponse, obj))
			}
		})
	}
}
