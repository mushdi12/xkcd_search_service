package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"yadro.com/course/api/core"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error { return f.err }

type fakeAuther struct {
	token string
	err   error
}

func (f fakeAuther) Login(user, password string) (string, error) {
	return f.token, f.err
}

type fakeUpdater struct {
	updateErr error
	stats     core.UpdateStats
	statsErr  error
	status    core.UpdateStatus
	statusErr error
	dropErr   error
}

func (f fakeUpdater) Update(ctx context.Context) error                    { return f.updateErr }
func (f fakeUpdater) Stats(ctx context.Context) (core.UpdateStats, error) { return f.stats, f.statsErr }
func (f fakeUpdater) Status(ctx context.Context) (core.UpdateStatus, error) {
	return f.status, f.statusErr
}
func (f fakeUpdater) Drop(ctx context.Context) error { return f.dropErr }

type fakeSearcher struct {
	comics []core.Comics
	err    error
}

func (f fakeSearcher) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	return f.comics, f.err
}

func (f fakeSearcher) SearchIndex(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	return f.comics, f.err
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewPingHandler(t *testing.T) {
	log := newTestLogger()
	pingers := map[string]core.Pinger{
		"s1": fakePinger{},
		"s2": fakePinger{err: errors.New("down")},
	}
	h := NewPingHandler(log, pingers)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp PingResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Replies["s1"] != "ok" || resp.Replies["s2"] != "unavailable" {
		t.Fatalf("unexpected replies: %#v", resp.Replies)
	}
}

func TestNewLoginHandler_BadJSON(t *testing.T) {
	log := newTestLogger()
	h := NewLoginHandler(log, fakeAuther{})

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestNewLoginHandler_AuthError(t *testing.T) {
	log := newTestLogger()
	h := NewLoginHandler(log, fakeAuther{err: errors.New("bad creds")})

	body, _ := json.Marshal(LoginRequest{Name: "u", Password: "p"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestNewLoginHandler_Success(t *testing.T) {
	log := newTestLogger()
	h := NewLoginHandler(log, fakeAuther{token: "token"})

	body, _ := json.Marshal(LoginRequest{Name: "u", Password: "p"})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "token" {
		t.Fatalf("expected token in body, got %q", rr.Body.String())
	}
}

func TestNewUpdateHandler_Errors(t *testing.T) {
	log := newTestLogger()

	hAccepted := NewUpdateHandler(log, fakeUpdater{updateErr: core.ErrAlreadyExists})
	rr := httptest.NewRecorder()
	hAccepted(rr, httptest.NewRequest(http.MethodPost, "/api/update", nil))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	hInternal := NewUpdateHandler(log, fakeUpdater{updateErr: errors.New("err")})
	rr = httptest.NewRecorder()
	hInternal(rr, httptest.NewRequest(http.MethodPost, "/api/update", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestNewUpdateHandler_Success(t *testing.T) {
	log := newTestLogger()
	h := NewUpdateHandler(log, fakeUpdater{})

	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodPost, "/api/update", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestNewUpdateStatsHandler(t *testing.T) {
	log := newTestLogger()
	stats := core.UpdateStats{WordsTotal: 1, WordsUnique: 2, ComicsFetched: 3, ComicsTotal: 4}

	h := NewUpdateStatsHandler(log, fakeUpdater{stats: stats})
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/api/update/stats", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp UpdateStats
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ComicsTotal != 4 {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestNewUpdateStatusHandler(t *testing.T) {
	log := newTestLogger()
	h := NewUpdateStatusHandler(log, fakeUpdater{status: core.StatusUpdateRunning})

	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/api/update/status", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp UpdateStatus
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != string(core.StatusUpdateRunning) {
		t.Fatalf("unexpected status: %#v", resp)
	}
}

func TestNewDropHandler(t *testing.T) {
	log := newTestLogger()
	h := NewDropHandler(log, fakeUpdater{})
	rr := httptest.NewRecorder()

	h(rr, httptest.NewRequest(http.MethodDelete, "/api/db", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestNewSearchHandler_BadLimitAndNoPhrase(t *testing.T) {
	log := newTestLogger()
	h := NewSearchHandler(log, fakeSearcher{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/search?limit=abc", nil)
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/search", nil)
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestNewSearchHandler_NotFoundAndSuccess(t *testing.T) {
	log := newTestLogger()

	hNF := NewSearchHandler(log, fakeSearcher{err: core.ErrNotFound})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/search?phrase=linux", nil)
	hNF(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	hOK := NewSearchHandler(log, fakeSearcher{
		comics: []core.Comics{{ID: 1, URL: "u1"}},
	})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/search?phrase=linux&limit=1", nil)
	hOK(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp ComicsReply
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Total != 1 || len(resp.Comics) != 1 {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

func TestNewIndexSearchHandler_Success(t *testing.T) {
	log := newTestLogger()
	h := NewIndexSearchHandler(log, fakeSearcher{
		comics: []core.Comics{{ID: 2, URL: "u2"}},
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/isearch?phrase=linux", nil)
	h(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp ComicsReply
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Total != 1 || resp.Comics[0].ID != 2 {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}
