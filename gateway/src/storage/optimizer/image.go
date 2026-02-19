package optimizer

import (
	"bytes"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/storage-gateway/src/storage"
)

func OptimizeImage(object *storage.PutObject) (*storage.PutObject, error) {
	contentType := object.ContentType
	if contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/png" {
		return object, nil
	}

	img, err := vips.NewImageFromReader(object.Body)

	if err != nil {
		return nil, err
	}
	var data []byte

	if img.HasAlpha() {
		data, _, err = img.ExportPng(&vips.PngExportParams{
			StripMetadata: true,
			Quality:       75,
			Interlace:     false,
			Compression:   8,
		})
		contentType = "image/png"
	} else {
		data, _, err = img.ExportJpeg(&vips.JpegExportParams{
			StripMetadata:  true,
			Quality:        75,
			Interlace:      true,
			OptimizeCoding: true,
		})
		contentType = "image/jpeg"
	}
	if err != nil {
		return nil, err
	}

	if object.Metadata == nil {
		object.Metadata = map[string]string{}
	}
	object.Metadata["optimized"] = "true"
	return &storage.PutObject{
		ContentType:   contentType,
		Body:          bytes.NewReader(data),
		Metadata:      object.Metadata,
		ContentLength: int64(len(data)),
	}, nil
}
