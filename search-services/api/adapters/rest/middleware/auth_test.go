package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeVerifier struct {
	err error
}

func (f fakeVerifier) Verify(token string) error {
	return f.err
}

func TestAuth_NoHeader(t *testing.T) {
	called := false
	h := Auth(func(http.ResponseWriter, *http.Request) {
		called = true
	}, fakeVerifier{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if called {
		t.Fatalf("handler should not be called")
	}
}

func TestAuth_BadPrefix(t *testing.T) {
	called := false
	h := Auth(func(http.ResponseWriter, *http.Request) {
		called = true
	}, fakeVerifier{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if called {
		t.Fatalf("handler should not be called")
	}
}

func TestAuth_EmptyToken(t *testing.T) {
	called := false
	h := Auth(func(http.ResponseWriter, *http.Request) {
		called = true
	}, fakeVerifier{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token ")
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if called {
		t.Fatalf("handler should not be called")
	}
}

func TestAuth_VerifyError(t *testing.T) {
	called := false
	h := Auth(func(http.ResponseWriter, *http.Request) {
		called = true
	}, fakeVerifier{err: assertAnError{}})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token token")
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if called {
		t.Fatalf("handler should not be called")
	}
}

type assertAnError struct{}

func (assertAnError) Error() string { return "err" }

func TestAuth_Success(t *testing.T) {
	called := false
	h := Auth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}, fakeVerifier{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token token")
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	if !called {
		t.Fatalf("handler should be called")
	}
}


