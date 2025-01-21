package middleware

import (
	"Backend/internal/database"
	"context"
	"net/http"
)

var CtxDbKey = struct{}{}

func ApplyAttachDb(db *database.GormPgAdapter) ApplyMiddlewareLayer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), CtxDbKey, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetDbFromContext(ctx context.Context) (*database.GormPgAdapter, bool) {
	db, ok := ctx.Value(CtxDbKey).(*database.GormPgAdapter)
	return db, ok
}
