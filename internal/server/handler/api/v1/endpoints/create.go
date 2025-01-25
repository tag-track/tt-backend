package endpoints

import (
	"Backend/internal/models"
	"Backend/internal/server/middleware"
	"Backend/internal/thumbnail"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucsky/cuid"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
)

func Create(w http.ResponseWriter, r *http.Request) {

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

	if id == "" {
		id = cuid.New()
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	parentId := r.FormValue("parent_id")

	// Extract images
	images := r.MultipartForm.File["images"]
	thumbnails := make([]string, 0)

	// Get Minio Object
	objStore, ok := middleware.GetObjStoreFromContext(r.Context())
	if !ok {
		http.Error(w, "Unable to load ObjectStore instance", http.StatusInternalServerError)
		return
	}

	var thErrFlag error
	mut := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(images))
	for _, img := range images {
		go func(img *multipart.FileHeader, _id string) {
			imgBody, _ := img.Open()
			t, err := thumbnail.NewThumbnailsFromMultipart(imgBody, _id)
			if err != nil {
				log.Printf("%v", err)
				thErrFlag = err
				return
			}

			if err := objStore.UploadThumbnail(r.Context(), t); err != nil {
				thErrFlag = err
				log.Printf("%v", err)
				return
			}

			mut.Lock()
			thumbnails = append(thumbnails, fmt.Sprintf("/images/v1/%s", t.GetImageBaseNameWithExt()))
			mut.Unlock()

			wg.Done()
		}(img, id)
	}
	wg.Wait()

	if thErrFlag != nil {
		http.Error(w, "Unable to generate thumbnails", http.StatusInternalServerError)
		return
	}

	entity := models.NewEntity(
		models.EntityWithId(id),
		models.EntityWithName(name),
		models.EntityWithDescription(description),
		models.EntityWithParentId(parentId),
		models.EntityWithImages(thumbnails),
	)

	// Create entity in the database
	db, ok := middleware.GetDbFromContext(r.Context())
	if !ok {
		http.Error(w, "Unable to load DB instance", http.StatusInternalServerError)
		return
	}
	if err := db.CreateEntity(r.Context(), entity); err != nil {
		log.Println(err)
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
