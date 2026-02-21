package optimizer

import (
	"strings"

	"github.com/storage-gateway/src/storage"
)

func Optimize(object *storage.PutObject) (*storage.PutObject, error) {
	if object.ContentLength < 500*1024 {
		return object, nil
	}
	if object.Metadata != nil && object.Metadata["optimized"] == "true" {
		return object, nil
	}

	if strings.HasPrefix(object.ContentType, "image/") {
		return OptimizeImage(object)
	}
	if strings.HasPrefix(object.ContentType, "video/") {
		return OptimizeVideo(object)
	}
	return object, nil
}
