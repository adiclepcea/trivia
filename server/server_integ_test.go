// +build integration

package server

import "testing"

func TestPopulateCategories(t *testing.T) {
	c, err := populateCategories(categoriesURL)

	if err != nil {
		t.Fatalf("Expected no error while fetching categories. Got: %s", err.Error())
	}

	if len(c) == 0 {
		t.Fatalf("Expected to get some categorie, but got 0")
	}
}
