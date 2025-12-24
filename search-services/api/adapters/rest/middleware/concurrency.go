package middleware

import (
	"net/http"
)

func Concurrency(next http.HandlerFunc, limit int) http.HandlerFunc {

	semaphore := make(chan struct{}, limit)

	return func(w http.ResponseWriter, r *http.Request) {

		select {
		case semaphore <- struct{}{}:
			next(w, r)
			<-semaphore
		default:
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
	}
}
