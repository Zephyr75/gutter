package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"
)

// Decoding an image file is expensive and the widget tree asks for the same file
// on every frame, so the result is cached by path. Resampling is no longer done
// here at all: an image is uploaded once at its native size and scaled by the
// sampler, so there is nothing to resize and nothing to re-cache when a widget
// changes size.

var (
	mu sync.Mutex

	decoded    = map[string]image.Image{}
	decodedErr = map[string]error{}
)

// GetImageFromFilePath decodes the file at filePath, or replays the result of
// the first attempt if it has been decoded before. Failures are cached too, so
// a missing asset costs one failed open rather than one per frame.
func GetImageFromFilePath(filePath string) (image.Image, error) {
	mu.Lock()
	defer mu.Unlock()

	if img, ok := decoded[filePath]; ok {
		return img, decodedErr[filePath]
	}

	f, err := os.Open(filePath)
	if err != nil {
		decoded[filePath], decodedErr[filePath] = nil, err
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	decoded[filePath], decodedErr[filePath] = img, err
	return img, err
}

// ClearImageCache drops every cached decode. Only useful for tests and for
// reloading assets that changed on disk.
func ClearImageCache() {
	mu.Lock()
	defer mu.Unlock()
	decoded = map[string]image.Image{}
	decodedErr = map[string]error{}
}
