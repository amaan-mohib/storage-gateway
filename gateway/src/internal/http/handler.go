package http

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/storage-gateway/src/internal/config"
	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/internal/storage"
	"github.com/storage-gateway/src/internal/storage/backup"
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

	// compress and upload

	err = h.files.Upload(r.Context(), bucket, key, file, storage.PutOptions{
		ContentType:   contentType,
		Metadata:      metadata,
		ContentLength: header.Size,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// queue this
	backup.ProcessBackup(backup.BackupJob{
		Key:    key,
		Bucket: bucket,
	})

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	exists := h.files.Exists(r.Context(), bucket, key)

	if exists {
		out, err := h.files.GetFile(r.Context(), bucket, key)
		if err != nil {
			log.Printf("exist nf: %s", err.Error())
			http.NotFound(w, r)
			return
		}
		defer out.Body.Close()

		io.Copy(w, out.Body)
	} else {
		out, err := backup.FetchFromBackup(r.Context(), backup.BackupJob{Key: key, Bucket: bucket})
		if err != nil {
			log.Printf("not exist nf: %s", err.Error())
			http.NotFound(w, r)
			return
		}
		defer out.Body.Close()

		io.Copy(w, out.Body)
	}
}
