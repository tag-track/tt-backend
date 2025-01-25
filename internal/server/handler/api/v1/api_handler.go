package v1

import "net/http"
import "Backend/internal/server/handler/api/v1/endpoints"

func Router() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("POST /create", http.HandlerFunc(endpoints.Create))
	router.HandleFunc("GET /query", http.HandlerFunc(endpoints.Query))

	return router
}
