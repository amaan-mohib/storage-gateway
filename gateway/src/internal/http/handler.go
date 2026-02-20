package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/backup"
)

type Handler struct {
	files *service.FileService
}

func NewHandler(files *service.FileService) *Handler {
	return &Handler{files: files}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
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

	putOptions := &storage.PutOptions{
		ContentType:   contentType,
		Metadata:      metadata,
		ContentLength: header.Size,
	}

	err = h.files.Upload(r.Context(), bucket, key, file, putOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queue.EnqueueBackup(queue.BackupJob{
		Key:    key,
		Bucket: bucket,
	})

	res, err := json.Marshal(putOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")
	ctx := r.Context()

	exists := h.files.Exists(ctx, bucket, key)

	if exists {
		out, err := h.files.GetFile(ctx, bucket, key)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer out.Body.Close()
		w.Header().Set("Content-Type", out.ContentType)

		io.Copy(w, out.Body)
	} else {
		out, err := backup.FetchFromBackup(ctx, &queue.BackupJob{Key: key, Bucket: bucket})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer out.Body.Close()

		w.Header().Set("Content-Type", out.ContentType)
		io.Copy(w, out.Body)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")
	deleteBackup := r.URL.Query().Get("deleteBackup")
	ctx := r.Context()

	err := h.files.Delete(ctx, bucket, key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if deleteBackup == "true" {
		queue.EnqueueDelete(queue.DeleteJob{
			Key:    key,
			Bucket: bucket,
		})
	}
	w.WriteHeader(http.StatusNoContent)
}
