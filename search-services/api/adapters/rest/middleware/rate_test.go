package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRate_NonPositiveRPSReturnsSameHandler(t *testing.T) {
	h1 := func(w http.ResponseWriter, r *http.Request) {}
	h2 := Rate(h1, 0)
	if h2 == nil {
		t.Fatalf("expected handler, got nil")
	}
}

func TestRate_AllowedRequest(t *testing.T) {
	called := false
	h := Rate(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}, 10)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Fatalf("handler was not called")
	}
}


