package v1

import "net/http"

func Router() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("POST /create", http.HandlerFunc(create))

	return router
}
