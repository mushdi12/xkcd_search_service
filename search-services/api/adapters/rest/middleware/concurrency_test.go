package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConcurrency_LimitOne(t *testing.T) {
	var (
		inHandler sync.WaitGroup
		done      sync.WaitGroup
	)

	inHandler.Add(1)
	done.Add(2)

	handler := Concurrency(func(w http.ResponseWriter, r *http.Request) {
		inHandler.Done()
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		done.Done()
	}, 1)

	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	}()

	inHandler.Wait()

	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)
		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", rr.Code)
		}
		done.Done()
	}()

	done.Wait()
}
