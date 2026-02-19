package optimizer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/storage-gateway/src/internal/storage"
)

func OptimizeVideo(object *storage.PutObject) (*storage.PutObject, error) {
	now := time.Now().Unix()
	inFile, err := os.CreateTemp("", fmt.Sprintf("input-%d-*.mp4", now))
	if err != nil {
		return nil, err
	}
	defer os.Remove(inFile.Name())
	defer inFile.Close()

	inData, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, err
	}

	if _, err := inFile.Write(inData); err != nil {
		return nil, err
	}

	outFile, err := os.CreateTemp("", fmt.Sprintf("output-%d-*.mp4", now))
	if err != nil {
		return nil, err
	}
	defer os.Remove(outFile.Name())
	defer outFile.Close()

	cmd := exec.Command("ffmpeg", "-y",
		"-i", inFile.Name(),
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "26",
		"-movflags", "frag_keyframe+empty_moov",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		outFile.Name(),
	)

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return nil, err
	}

	size := int64(len(data))
	if size >= object.ContentLength {
		return object, nil
	}

	obj := &storage.PutObject{
		Body:          bytes.NewReader(data),
		ContentType:   "video/mp4",
		ContentLength: size,
	}
	if obj.Metadata == nil {
		obj.Metadata = map[string]string{}
	}
	obj.Metadata["optimized"] = "true"
	return obj, nil
}
