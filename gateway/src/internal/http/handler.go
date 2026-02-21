package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/processing"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
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
	ctx := r.Context()

	if h.files.Exists(ctx, bucket, key) {
		http.Error(w, "Key already exists", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	contentType := header.Header.Get("Content-Type")
	isImageOrVideo := strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
	if !isImageOrVideo {
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		contentType = http.DetectContentType(buffer[:n])
		file.Seek(0, io.SeekStart)
	}

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

	err = h.files.Upload(ctx, bucket, key, file, putOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	queue.EnqueueBackup(queue.BackupJob{
		Key:    key,
		Bucket: bucket,
	})

	if strings.HasPrefix(contentType, "video/") {
		queue.EnqueueGenerateThumb(queue.GenerateThumbJob{
			Key:    key,
			Bucket: bucket,
		})
	}

	res, err := json.Marshal(putOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func writeCacheHeaders(w http.ResponseWriter, r *http.Request, file *storage.GetObject, tempCache bool) bool {
	w.Header().Set("Content-Type", file.ContentType)
	if file.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(file.ContentLength, 10))
	}
	w.Header().Set("ETag", file.ETag)
	w.Header().Set("Last-Modified", file.LastModified.Format(http.TimeFormat))
	if tempCache {
		w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400, stale-if-error=1200")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, stale-if-error=1200, immutable")
	}

	if match := r.Header.Get("If-None-Match"); match == file.ETag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		t, err := http.ParseTime(ifModifiedSince)
		if err == nil && file.LastModified.Before(t) {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}
	return false
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

		if ret := writeCacheHeaders(w, r, out, false); ret {
			return
		}

		io.Copy(w, out.Body)
	} else {
		var out *storage.GetObject
		var err error
		payload := &queue.BackupJob{Key: key, Bucket: bucket}
		isThumb := strings.HasSuffix(key, processing.ThumbExt)

		if isThumb {
			out, err = processing.FetchAndGenerateThumb(ctx, h.files, payload)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		} else {
			out, err = processing.FetchFromBackup(ctx, payload)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer out.Body.Close()

		if ret := writeCacheHeaders(w, r, out, !isThumb); ret {
			return
		}

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
