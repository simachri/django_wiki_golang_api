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
	"strconv"
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

// Create a root and a child article and fetch the child article by its ID.
func TestGetArticleById(t *testing.T) {
	// Read the environment variables for the DB connection.
	godotenv.Load("../../.env")
	// Override the database name to use the testing database.
	os.Setenv("PGDATABASE", "go_api_tests")

	clearDB()

	// Create the following article hierarchy:
	// /  (root)
	// /unit1
	router := setupRouter()

	// Create the root article.
	body, err := json.Marshal(m.RootArticle{ArticleBase: m.ArticleBase{
		Title:   "Root article created from unit test",
		Content: "# First header",
	},
	})
	assert.Nil(t, err)
	req, err := http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Unmarshal the root article response to get its ID.
	var root m.RootArticle
	err = json.Unmarshal([]byte(w.Body.String()), &root)
	assert.Nil(t, err)

	// Create the child artile /unit1 under root /.
	var art1 = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
		},
		Slug: "unit1",
	}
	body, err = json.Marshal(art1)
	assert.Nil(t, err)
	req, err = http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// Creating a new recorder is required to avoid side effects with previous HTTP
	// calls.
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "expected return code %v, but got %v", http.StatusCreated, w.Code)
	// fmt.Printf("Response body: %v\n", w.Body.String())
	var resArt1 m.Article
	err = json.Unmarshal([]byte(w.Body.String()), &resArt1)
	assert.Nil(t, err)

    // TEST
    // Select the child article /unit1
	req, err = http.NewRequest(http.MethodGet, "/articles/" + strconv.Itoa(resArt1.ID), nil)
	assert.Nil(t, err)
	// Creating a new recorder is required to avoid side effects with previous HTTP
	// calls.
	w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "expected return code %v, but got %v", http.StatusOK, w.Code)
	var res m.Article
	err = json.Unmarshal([]byte(w.Body.String()), &res)
	assert.Nil(t, err)
	var artExp = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
			Left:        2,
			Right:       3,
		},
		Slug:  "unit1",
		Level: 1,
	}
	assert.Equal(t, artExp.Title, res.Title, "Title differs")
	assert.Equal(t, artExp.Content, res.Content, "Content differs")
	assert.Equal(t, artExp.Slug, res.Slug, "Slug differs")
	assert.Equal(t, artExp.Level, res.Level, "Level differs")
	assert.Equal(t, artExp.ParentArtID, res.ParentArtID, "ParentArtID differs")
	assert.Equal(t, artExp.Left, res.Left, "Left differs")
	assert.Equal(t, artExp.Right, res.Right, "Right differs")
}

// Create a second article that is a child of the root article and a sibling to an
// existing child article.
func TestAddSecondChildToRoot(t *testing.T) {
	// Read the environment variables for the DB connection.
	godotenv.Load("../../.env")
	// Override the database name to use the testing database.
	os.Setenv("PGDATABASE", "go_api_tests")

	clearDB()

	// Create the following article hierarchy:
	// /  (root)
	// /unit1
	// /unit2
	router := setupRouter()

	// Create the root article.
	body, err := json.Marshal(m.RootArticle{ArticleBase: m.ArticleBase{
		Title:   "Root article created from unit test",
		Content: "# First header",
	},
	})
	assert.Nil(t, err)
	req, err := http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Unmarshal the root article response to get its ID.
	var root m.RootArticle
	err = json.Unmarshal([]byte(w.Body.String()), &root)
	assert.Nil(t, err)

	// Create the first child artile /unit1 under root /.
	var art1 = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
		},
		Slug: "unit1",
	}
	body, err = json.Marshal(art1)
	assert.Nil(t, err)
	req, err = http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// Creating a new recorder is required to avoid side effects with previous HTTP
	// calls.
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "expected return code %v, but got %v", http.StatusCreated, w.Code)
	// fmt.Printf("Response body: %v\n", w.Body.String())
	var resArt1 m.Article
	err = json.Unmarshal([]byte(w.Body.String()), &resArt1)
	assert.Nil(t, err)

	// TEST
	// Create the second child article /unit2 under root /..
	var art = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Second child article under root",
			Content:     "# Foo Bar Baz",
			ParentArtID: root.ID,
		},
		Slug: "unit2",
	}
	body, err = json.Marshal(art)
	assert.Nil(t, err)
	req, err = http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// Creating a new recorder is required to avoid side effects with previous HTTP
	// calls.
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

    // Child article validation: /unit2
	assert.Equal(t, http.StatusCreated, w.Code, "expected return code %v, but got %v", http.StatusCreated, w.Code)
	// fmt.Printf("Response body: %v\n", w.Body.String())
	var resArt2 m.Article
	err = json.Unmarshal([]byte(w.Body.String()), &resArt2)
	assert.Nil(t, err)
	var artExp = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Second child article under root",
			Content:     "# Foo Bar Baz",
			ParentArtID: root.ID,
			Left:        4,
			Right:       5,
		},
		Slug:  "unit2",
		Level: 1,
	}
	assert.Equal(t, artExp.Title, resArt2.Title, "Article unit2: Title differs")
	assert.Equal(t, artExp.Content, resArt2.Content, "Article unit2: Content differs")
	assert.Equal(t, artExp.Slug, resArt2.Slug, "Article unit2: Slug differs")
	assert.Equal(t, artExp.Level, resArt2.Level, "Article unit2: Level differs")
	assert.Equal(t, artExp.ParentArtID, resArt2.ParentArtID, "Article unit2: ParentArtID differs")
	assert.Equal(t, artExp.Left, resArt2.Left, "Article unit2: Left differs")
	assert.Equal(t, artExp.Right, resArt2.Right, "Article unit2: Right differs")

    // Child article validation: /unit1 must not have been changed.
	artExp = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
			Left:        2,
			Right:       3,
		},
		Slug:  "unit1",
		Level: 1,
	}
	assert.Equal(t, artExp.Title, resArt1.Title, "Article unit1: Title differs")
	assert.Equal(t, artExp.Content, resArt1.Content, "Article unit1: Content differs")
	assert.Equal(t, artExp.Slug, resArt1.Slug, "Article unit1: Slug differs")
	assert.Equal(t, artExp.Level, resArt1.Level, "Article unit1: Level differs")
	assert.Equal(t, artExp.ParentArtID, resArt1.ParentArtID, "Article unit1: ParentArtID differs")
	assert.Equal(t, artExp.Left, resArt1.Left, "Article unit1: Left differs")
	assert.Equal(t, artExp.Right, resArt1.Right, "Article unit1: Right differs")

	// Root article validation: Make sure that the 'Right' proprerty of the root article has been adjusted.
	req, err = http.NewRequest("GET", "/articles", nil)
	assert.Nil(t, err)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &root)
	assert.Nil(t, err)
	rootExp := m.RootArticle{ArticleBase: m.ArticleBase{
		Title:   "Root article created from unit test",
		Content: "# First header",
		Left:    1,
		Right:   6,
	},
	}
	assert.Equal(t, rootExp.Title, root.Title, "Root: Title differs")
	assert.Equal(t, rootExp.Content, root.Content, "Root: Content differs")
	assert.Equal(t, rootExp.ParentArtID, root.ParentArtID, "Root: ParentArtID differs")
	assert.Equal(t, rootExp.Left, root.Left, "Root: Left differs")
	assert.Equal(t, rootExp.Right, root.Right, "Root: Right differs")
}

// Create an article that is a child of the root article.
func TestAddChildToRoot(t *testing.T) {
	// Read the environment variables for the DB connection.
	godotenv.Load("../../.env")
	// Override the database name to use the testing database.
	os.Setenv("PGDATABASE", "go_api_tests")

	clearDB()

	// Create the following article hierarchy:
	// /  (root)
	// /unit1
	router := setupRouter()

	// Create the root article.
	body, err := json.Marshal(m.RootArticle{ArticleBase: m.ArticleBase{
		Title:   "Root article created from unit test",
		Content: "# First header",
	},
	})
	assert.Nil(t, err)
	req, err := http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Unmarshal the root article response to get its ID.
	var root m.RootArticle
	err = json.Unmarshal([]byte(w.Body.String()), &root)
	assert.Nil(t, err)

	// TEST
	// Create the child article under root.
	var art = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
		},
		Slug: "unit1",
	}
	body, err = json.Marshal(art)
	assert.Nil(t, err)
	req, err = http.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(body))
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// Creating a new recorder is required to avoid side effects with previous HTTP
	// calls.
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Child article validation.
	assert.Equal(t, http.StatusCreated, w.Code, "expected return code %v, but got %v", http.StatusCreated, w.Code)
	// fmt.Printf("Response body: %v\n", w.Body.String())
	var res m.Article
	err = json.Unmarshal([]byte(w.Body.String()), &res)
	assert.Nil(t, err)
	var artExp = m.Article{
		ArticleBase: m.ArticleBase{
			Title:       "Child article below root",
			Content:     "# Child article header",
			ParentArtID: root.ID,
			Left:        2,
			Right:       3,
		},
		Slug:  "unit1",
		Level: 1,
	}
	assert.Equal(t, artExp.Title, res.Title, "Title differs")
	assert.Equal(t, artExp.Content, res.Content, "Content differs")
	assert.Equal(t, artExp.Slug, res.Slug, "Slug differs")
	assert.Equal(t, artExp.Level, res.Level, "Level differs")
	assert.Equal(t, artExp.ParentArtID, res.ParentArtID, "ParentArtID differs")
	assert.Equal(t, artExp.Left, res.Left, "Left differs")
	assert.Equal(t, artExp.Right, res.Right, "Right differs")

	// Root article validation: Make sure that the 'Right' proprerty of the root article has been adjusted.
	req, err = http.NewRequest("GET", "/articles", nil)
	assert.Nil(t, err)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &root)
	assert.Nil(t, err)
	rootExp := m.RootArticle{ArticleBase: m.ArticleBase{
		Title:   "Root article created from unit test",
		Content: "# First header",
		Left:    1,
		Right:   4,
	},
	}
	assert.Equal(t, rootExp.Title, root.Title, "Title differs")
	assert.Equal(t, rootExp.Content, root.Content, "Content differs")
	assert.Equal(t, rootExp.ParentArtID, root.ParentArtID, "ParentArtID differs")
	assert.Equal(t, rootExp.Left, root.Left, "Left differs")
	assert.Equal(t, rootExp.Right, root.Right, "Right differs")
}

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
		{"GET root article", "GET", "/articles", nil, http.StatusOK, "",
			&m.RootArticle{
				ArticleBase: m.ArticleBase{
					Title:   "Root article created from testing",
					Content: "# Hello World"}}},
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
				// fmt.Printf("Response body: %v\n", w.Body.String())
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
