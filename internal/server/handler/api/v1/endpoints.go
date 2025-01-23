package v1

import (
	"Backend/internal/models"
	"Backend/internal/server/middleware"
	"Backend/internal/thumbnail"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
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
	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	parentId := r.FormValue("parent_id")

	// Extract images
	images := r.MultipartForm.File["images"]
	wg := sync.WaitGroup{}
	wg.Add(len(images))
	for _, img := range images {
		go func(img *multipart.FileHeader) {
			imgBody, _ := img.Open()
			_, _ = thumbnail.NewThumbnailsFromMultipart(imgBody)
			wg.Done()
		}(img)
	}
	wg.Wait()

	entity := models.NewEntity(
		models.EntityWithId(id),
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

func query(w http.ResponseWriter, r *http.Request) {

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
