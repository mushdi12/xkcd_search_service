package xkcd

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yadro.com/course/update/core"
)

func newTestClient(baseURL string) *Client {
	logger := slog.Default()
	return &Client{
		log: logger,
		client: http.Client{
			Timeout: time.Second,
		},
		url: baseURL,
	}
}

func TestClient_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"num":        42,
			"img":        "http://example.com/img.png",
			"title":      "t",
			"safe_title": "st",
			"transcript": "tr",
			"alt":        "alt",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)

	info, err := c.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if info.ID != 42 || info.URL == "" {
		t.Fatalf("unexpected info: %#v", info)
	}
	if info.Description == "" {
		t.Fatalf("expected non-empty description")
	}
}

func TestClient_LastID_UsesGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"num": 99}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)

	id, err := c.LastID(context.Background())
	if err != nil {
		t.Fatalf("LastID returned error: %v", err)
	}
	if id != 99 {
		t.Fatalf("expected 99, got %d", id)
	}
}

func TestClient_Get_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)

	_, err := c.Get(context.Background(), 1)
	if err != core.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
