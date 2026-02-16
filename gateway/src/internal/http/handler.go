package http

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/storage-gateway/src/internal/config"
	"github.com/storage-gateway/src/internal/service"
)

type Handler struct {
	files *service.FileService
}

func NewHandler(files *service.FileService) *Handler {
	return &Handler{files: files}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-AccessToken")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	decodedToken, decodeErr := base64.StdEncoding.DecodeString(token)
	if decodeErr != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	if string(decodedToken) != config.GetSafeEnv("ADMIN_ACCESS_TOKEN", "admin123") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	file, header, err := r.FormFile("file")
	contentType := header.Header.Get("Content-Type")

	var metadata map[string]string
	metadataStr := r.FormValue("metadata")
	if metadataStr != "" {
		err = json.Unmarshal([]byte(metadataStr), &metadata)
		if err != nil {
			http.Error(w, "Invalid metadata JSON", http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = h.files.Upload(r.Context(), bucket, key, file, contentType, metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// backup to specified service if enabled

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	exists := h.files.Exists(r.Context(), bucket, key)

	if exists {
		body, err := h.files.GetFile(r.Context(), bucket, key)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer body.Close()

		io.Copy(w, body)
	} else {
		// try to fetch from backup service if enabled and put it in primary storage for future requests

		http.NotFound(w, r)
	}
}
