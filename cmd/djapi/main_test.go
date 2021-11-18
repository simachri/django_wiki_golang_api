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

	m "coco-life.de/wapi/internal/models"
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

//// Create an article that is a child of the root article.
//func TestAddChildToRoot(t *testing.T) {
	//// Read the environment variables for the DB connection.
	//godotenv.Load("../../.env")
	//// Override the database name to use the testing database.
	//os.Setenv("PGDATABASE", "go_api_tests")

	//clearDB()

	//// Create the following article hierarchy:
	//// /  (root)
	//// /unit1
	//router := setupRouter()
	//w := httptest.NewRecorder()

	//requestBody, err := json.Marshal(m.RootArticle{ArticleBase: m.ArticleBase{
		//Title:   "Root article created from unit test",
		//Content: "# First header"},
	//})
	//assert.Nil(t, err)
	//req, _ := http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(requestBody))
	//req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	//router.ServeHTTP(w, req)
	//// Get the root article ID.
	//var root m.RootArticle
	//err = json.Unmarshal([]byte(w.Body.String()), &root)
	//assert.Nil(t, err)

	//var art = m.Article{
		//ArticleBase: m.ArticleBase{
			//Title: "Child article below root",
			//Content: "# Child article header",
		//},
		//ParentID: root.ID,
		//Slug:     "unit1",
	//}
	//requestBody, err = json.Marshal(art)
	//assert.Nil(t, err)
	//req, _ = http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(requestBody))
	//req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	//router.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusOK, w.Code, "expected return code %v, but got %v", http.StatusOK, w.Code)
	//var res m.Article
	//err = json.Unmarshal([]byte(w.Body.String()), &res)
	//assert.Nil(t, err)
	//assert.True(t, art.Equals(res),
		//fmt.Sprintf(
			//"JSON response differs.\n"+
				//"Exp: %v\n"+
				//"Act: %v\n", art, res))
//}

func TestBasics(t *testing.T) {
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
		{"Create root article", "POST", "/articles",
			&m.RootArticle{
				ArticleBase: m.ArticleBase{
					Title:   "Root article created from testing",
					Content: "# Hello World"}},
			http.StatusCreated, "", nil},
		//{"Create root article", "POST", "/articles",
		//&m.Article{
		//Title:   "Article created from testing",
		//Content: "# Hello World",
		//Slug:    "root",
		//ParentID: -1},
		//http.StatusCreated, "", nil},
		// The following test requires the POST /articles test to be run first.
		// FIXME: Make this test independent of POST /articles.
		{"GET root article", "GET", "/articles", nil, http.StatusOK, "",
			&m.RootArticle{
				ArticleBase: m.ArticleBase{
					Title:   "Root article created from testing",
					Content: "# Hello World"}}},
		//{"GET root article", "GET", "/articles/root", nil, http.StatusOK, "",
		//&m.Article{
		//Title:   "Article created from testing",
		//Content: "# Hello World",
		//Slug:    "root",
		//ParentID: -1}},
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
