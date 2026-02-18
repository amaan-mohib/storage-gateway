package optimizer

import (
	"strings"

	"github.com/storage-gateway/src/internal/storage"
)

func Optimize(object *storage.PutObject) (*storage.PutObject, error) {
	if object.Metadata["optimized"] == "true" {
		return object, nil
	}

	if strings.HasPrefix(object.ContentType, "image/") {
		return OptimizeImage(object)
	}
	return object, nil
}
