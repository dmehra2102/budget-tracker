package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/dmehra2102/budget-tracker/pkg/response"
)

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan bool)
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				response.Error(w, ctx.Err(), http.StatusRequestTimeout)
			}
		})
	}
}
