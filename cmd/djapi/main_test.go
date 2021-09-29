package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	m "coco-life.de/wapi/internal"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func clearDB() {
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	_, err = dbpool.Exec(context.Background(), "TRUNCATE wiki_article CASCADE;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to TRUNCATE wiki_article: %v\n", err)
		os.Exit(1)
	}
}

func TestApiRoutes(t *testing.T) {
	// Read the environment variables for the DB connection.
	godotenv.Load("../../.env")
	// Override the database name to use the testing database.
	os.Setenv("PGDATABASE", "go_api_tests")

	clearDB()

	cases := []struct {
		descr       string
		httpType    string
		endpoint    string
		bodyJSON    interface{}
		expCode     int
		expString   string
		expResponse m.Resource
	}{
		{"Ping API", "GET", "/ping", nil, http.StatusOK, "pong", nil},
		{"Database healthcheck", "GET", "/db/health", nil, http.StatusOK, "", nil},
		{"Create new article", "POST", "/articles",
			&m.Article{
				Title:   "Article created from testing",
				Content: "# Hello World",
				Slug:    "unit",
                ParentID: -1},
			http.StatusCreated, "", nil},
        {"GET article by slug", "GET", "/articles/unit", nil, http.StatusOK, "",
            &m.Article{
				Title:   "Article created from testing",
				Content: "# Hello World",
				Slug:    "unit",
                ParentID: -1}},
		//{"GET endpoint /articles/1", "GET", "/articles/1", nil, http.StatusOK, "",
			//&m.Article{
				//ID:      1,
				//Title:   "Article created from testing",
				//Content: "# Hello World"}},
	}

	router := setupRouter()
	for _, tc := range cases {
		t.Run(tc.descr, func(t *testing.T) {

			w := httptest.NewRecorder()
			requestBody, err := json.Marshal(tc.bodyJSON)
			assert.Nil(t, err)
			req, _ := http.NewRequest(tc.httpType, tc.endpoint, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json; charset=UTF-8")
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
				obj := data.(m.Resource)
				assert.True(t, tc.expResponse.Equals(obj),
					fmt.Sprintf(
						"JSON response differs.\n"+
							"Exp: %v\n"+
							"Act: %v\n", tc.expResponse, obj))
			}
		})
	}
}
