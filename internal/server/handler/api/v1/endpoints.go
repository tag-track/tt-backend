package v1

import (
	"Backend/internal/models"
	"Backend/internal/server/middleware"
	"encoding/json"
	"errors"
	"net/http"
)

func create(w http.ResponseWriter, r *http.Request) {

	// Parsing the request, assumes POST form-data
	const maxMemory = 10 << 20 // 10 MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		switch {
		case errors.Is(err, http.ErrNotMultipart) || errors.Is(err, http.ErrMissingBoundary):
			http.Error(w, "Form data not present or not multipart", http.StatusBadRequest)
		case err.Error() == "multipart: NextPart: EOF":
			http.Error(w, "No form data present", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to parse form, 10MB limit exceeded", http.StatusRequestEntityTooLarge)
		}
		return
	}

	// Extract form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	parentId := r.FormValue("parent_id")

	entity := models.NewEntity(
		models.EntityWithName(name),
		models.EntityWithDescription(description),
		models.EntityWithParentId(parentId),
	)

	// Create entity in the database
	db, ok := middleware.GetDbFromContext(r.Context())
	if !ok {
		http.Error(w, "Unable to load DB instance", http.StatusInternalServerError)
		return
	}
	if err := db.CreateEntity(r.Context(), entity); err != nil {
		http.Error(w, "Unable to insert entity", http.StatusInternalServerError)
		return
	}

	// Return created entity
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entity); err != nil {
		http.Error(w, "Unable to create response", http.StatusInternalServerError)
		return
	}

}
