package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO
// Post a new blog
// View blogs
// no need to mock gnats: https://gist.github.com/milosgajdos83/4d3e6bd6ed62744ea27e
func TestNewBlog(t *testing.T) {
	req, err := http.NewRequest("GET", "/new", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(newBlog)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
