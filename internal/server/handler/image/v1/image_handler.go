package v1

import (
	"Backend/internal/server/handler/image/v1/endpoints"
	"net/http"
)

func Router() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", endpoints.Index)

	return router
}
