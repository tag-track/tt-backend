package endpoints

import (
	"Backend/internal/models"
	"Backend/internal/server/middleware"
	"encoding/json"
	"log"
	"net/http"
)

func Query(w http.ResponseWriter, r *http.Request) {

	// Get ids params
	queryParams := r.URL.Query()
	ids := queryParams["id"]

	db, ok := middleware.GetDbFromContext(r.Context())
	if !ok {
		http.Error(w, "Unable to load DB instance", http.StatusInternalServerError)
		return
	}

	handleDbError := func(err error) {
		log.Printf("[Error] Unable to query DB, %v", err)
		http.Error(w, "Unable to fulfill request", http.StatusInternalServerError)
	}

	handleRes := func(entities []*models.Entity) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entities); err != nil {
			http.Error(w, "Unable to create response", http.StatusInternalServerError)
			return
		}
	}

	switch {
	case len(ids) == 0:
		entities, err := db.QueryTopLevel(r.Context())
		if err != nil {
			handleDbError(err)
		}
		handleRes(entities)
		return
	case len(ids) == 1:
		entity, err := db.QueryById(r.Context(), ids[0])
		if err != nil {
			handleDbError(err)
		}
		handleRes([]*models.Entity{entity})
		return
	case len(ids) > 1:
		entities, err := db.QueryMultipleById(r.Context(), ids...)
		if err != nil {
			handleDbError(err)
		}
		handleRes(entities)
		return
	}
}
