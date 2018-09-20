package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCategories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cr := CategoriesResponse{
			Categories: []Category{
				{
					Name: "Test1",
					ID:   1,
				},
				{
					Name: "Test2",
					ID:   2,
				},
			},
		}

		data, err := json.Marshal(&cr)
		if err != nil {
			t.Fatalf("Error creating response json in test: %s", err.Error())
		}

		w.Write(data)
	}))

	defer ts.Close()

	cats, err := populateCategories(ts.URL)

	if err != nil {
		t.Fatalf("No error expected when fetching categories, got %s", err.Error())
	}

	if len(cats) != 2 {
		t.Fatalf("Expected two categories, but got %d", len(cats))
	}
}
