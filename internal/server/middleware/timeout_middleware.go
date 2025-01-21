package middleware

import (
	"context"
	"net/http"
	"time"
)

func ApplyTimeout(duration time.Duration) ApplyMiddlewareLayer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-ctx.Done():
				http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			case <-done:
				return
			}
		})
	}
}
