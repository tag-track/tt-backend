package endpoints

import (
	"Backend/internal/server/middleware"
	"Backend/internal/thumbnail"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type imageDetails struct {
	Name string
	Size thumbnail.Size
}

func (i *imageDetails) toFullName() (string, error) {
	abvr, err := i.Size.Abvr()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s_%s.jpeg", i.Name, abvr), nil
}

func getImageDetails(path *url.URL) (*imageDetails, error) {

	// Handle empty path case
	if path.Path == "" {
		return nil, errors.New("empty path")
	}

	// Clean and split the path
	cleanPath := strings.Trim(path.Path, "/")
	seg := strings.Split(cleanPath, "/")

	if len(seg) != 1 {
		return nil, errors.New("malformed path")
	}

	filename := seg[0]
	if !strings.HasSuffix(filename, ".jpeg") {
		return nil, errors.New("image type not supported")
	}

	// Handle image name extraction
	imgName := strings.TrimSuffix(filename, ".jpeg")

	// Handle size parameter
	sizeParam := path.Query().Get("size")
	size, err := thumbnail.SizeFromAbvr(sizeParam)
	if err != nil || size == thumbnail.ErrorSize {
		size = thumbnail.MediumSize // Default size
	}

	return &imageDetails{
		Name: imgName,
		Size: size,
	}, nil
}

func Index(w http.ResponseWriter, r *http.Request) {

	log.Printf("%v", r.URL.Path)

	imgDetails, err := getImageDetails(r.URL)
	if err != nil {
		log.Printf("[Error] Unable to get image details: %v", err)
		http.Error(w, "Unable to parse image details", http.StatusInternalServerError)
		return
	}
	imgToRetrieve, err := imgDetails.toFullName()
	if err != nil {
		log.Printf("[Error] Unable to derive full image name: %v", err)
		http.Error(w, "Unable to derive full image name", http.StatusInternalServerError)
		return
	}

	objStore, ok := middleware.GetObjStoreFromContext(r.Context())
	if !ok {
		log.Printf("[Error] Unable to retrieve object store")
		http.Error(w, "Unable to retrieve object store", http.StatusInternalServerError)
		return
	}

	imgData, err := objStore.RetrieveImage(r.Context(), imgToRetrieve)
	if err != nil {
		log.Printf("[Error] Unable to retrieve object")
		http.Error(w, "Unable to retrieve object", http.StatusInternalServerError)
		return
	}
	defer imgData.Close()

	img, err := io.ReadAll(imgData)
	if err != nil {
		log.Printf("[Error] Unable to read object")
		http.Error(w, "Unable to read object", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(img)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(img)))
	w.Header().Set("Cache-Control", "public, max-age=86400") // one day, 60 * 60 * 24
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(img); err != nil {
		log.Printf("[Error] Unable write image")
		http.Error(w, "Unable to write image", http.StatusInternalServerError)
		return
	}

}
