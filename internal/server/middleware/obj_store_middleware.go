package middleware

import (
	"Backend/internal/objectstore"
	"context"
	"net/http"
)

func ApplyAttachObjStore(db *objectstore.MinioAdapter) ApplyMiddlewareLayer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ContextKeyObjStore, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetObjStoreFromContext(ctx context.Context) (*objectstore.MinioAdapter, bool) {
	db, ok := ctx.Value(ContextKeyObjStore).(*objectstore.MinioAdapter)
	return db, ok
}
