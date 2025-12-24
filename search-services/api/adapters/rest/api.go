package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/core"
)

// "GET /api/ping"
func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := PingResponse{
			Replies: make(map[string]string),
		}
		for name, pinger := range pingers {
			if err := pinger.Ping(context.Background()); err != nil {
				reply.Replies[name] = "unavailable"
				log.Error("one of services is not available", "service", name, "error", err)
				continue
			}
			reply.Replies[name] = "ok"
		}

		if err := encodeReply(w, reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

// "POST /api/login"
func NewLoginHandler(log *slog.Logger, auther core.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error("cannot decode request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := auther.Login(request.Name, request.Password)
		if err != nil {
			log.Error("failed to login", "error", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if _, err := w.Write([]byte(token)); err != nil {
			log.Error("cannot write token", "error", err)
		}
	}
}

// "POST /api/update"
func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			log.Error("error while updating", "error", err)
			if errors.Is(err, core.ErrAlreadyExists) {
				http.Error(w, err.Error(), http.StatusAccepted)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// "GET /api/update/stats"
func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("error while getting stats", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reply := UpdateStats{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}

		if err := encodeReply(w, reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

// "GET /api/update/status"
func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("error while getting status", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reply := UpdateStatus{
			Status: string(status),
		}

		if err := encodeReply(w, reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

// "DELETE /api/db"
func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("error while dropping", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// "GET /api/search"
func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var limit int
		var err error
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				log.Error("wrong limit", "value", limitStr)
				http.Error(w, "bad limit", http.StatusBadRequest)
				return
			}
			if limit < 0 {
				log.Error("wrong limit", "value", limit)
				http.Error(w, "bad limit", http.StatusBadRequest)
				return
			}
		}
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("no phrase")
			http.Error(w, "no phrase", http.StatusBadRequest)
			return
		}

		comics, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			if errors.Is(err, core.ErrNotFound) {
				http.Error(w, "no comics found", http.StatusNotFound)
				return
			}
			log.Error("error while seaching", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reply := ComicsReply{
			Comics: make([]Comics, 0, len(comics)),
			Total:  len(comics),
		}
		for _, c := range comics {
			reply.Comics = append(reply.Comics, Comics{ID: c.ID, URL: c.URL})
		}

		if err := encodeReply(w, reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

// "GET /api/isearch"
func NewIndexSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var limit int
		var err error
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				log.Error("wrong limit", "value", limitStr)
				http.Error(w, "bad limit", http.StatusBadRequest)
				return
			}
			if limit < 0 {
				log.Error("wrong limit", "value", limit)
				http.Error(w, "bad limit", http.StatusBadRequest)
				return
			}
		}
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("no phrase")
			http.Error(w, "no phrase", http.StatusBadRequest)
			return
		}

		comics, err := searcher.SearchIndex(r.Context(), phrase, limit)
		if err != nil {
			if errors.Is(err, core.ErrNotFound) {
				http.Error(w, "no comics found", http.StatusNotFound)
				return
			}
			log.Error("error while seaching", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reply := ComicsReply{
			Comics: make([]Comics, 0, len(comics)),
			Total:  len(comics),
		}
		for _, c := range comics {
			reply.Comics = append(reply.Comics, Comics{ID: c.ID, URL: c.URL})
		}

		if err := encodeReply(w, reply); err != nil {
			log.Error("cannot encode reply", "error", err)
		}
	}
}

func encodeReply(w io.Writer, reply any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(reply); err != nil {
		return fmt.Errorf("could not encode comics: %v", err)
	}
	return nil
}
