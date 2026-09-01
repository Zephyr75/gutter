package utils

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"github.com/nfnt/resize"
)

// Decoding a PNG and resampling it are the two most expensive things the
// rasteriser does, and the widget tree asks for the same file at the same size
// on every frame. Both results are cached: the decode by path, the resample by
// path and target size.

var (
	mu sync.Mutex

	decoded    = map[string]image.Image{}
	decodedErr = map[string]error{}

	scaled = map[string][]scaledImage{}
)

type scaledImage struct {
	width, height int
	img           image.Image
}

// Dragging a window edge asks for a different size every frame, so the variant
// list per path is bounded and evicted oldest-first rather than growing without
// limit.
const maxVariantsPerImage = 4

var (
	ErrZeroSize = errors.New("gutter: image requested at zero size")
	ErrNoImage  = errors.New("gutter: file decoded to a nil image")
)

// GetImageFromFilePath decodes the file at filePath, or replays the result of
// the first attempt if it has been decoded before. Failures are cached too, so
// a missing asset costs one failed open rather than one per frame.
func GetImageFromFilePath(filePath string) (image.Image, error) {
	mu.Lock()
	defer mu.Unlock()
	return decode(filePath)
}

// decode assumes mu is held.
func decode(filePath string) (image.Image, error) {
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

// GetScaledImage returns the file at filePath resampled to width x height,
// reusing the previous result whenever that size is asked for again. This is
// the call the draw path should use: it collapses an open, a decode and a
// Lanczos3 resample into a map lookup on every frame after the first.
func GetScaledImage(filePath string, width, height int) (image.Image, error) {
	if width <= 0 || height <= 0 {
		return nil, ErrZeroSize
	}

	mu.Lock()
	defer mu.Unlock()

	for _, v := range scaled[filePath] {
		if v.width == width && v.height == height {
			return v.img, nil
		}
	}

	src, err := decode(filePath)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, ErrNoImage
	}

	img := resize.Resize(uint(width), uint(height), src, resize.Lanczos3)

	list := append(scaled[filePath], scaledImage{width, height, img})
	if len(list) > maxVariantsPerImage {
		// Shift down rather than reslice, so the backing array stays put
		copy(list, list[1:])
		list = list[:maxVariantsPerImage]
	}
	scaled[filePath] = list

	return img, nil
}

// ClearImageCache drops every cached decode and resample. Only useful for tests
// and for reloading assets that changed on disk.
func ClearImageCache() {
	mu.Lock()
	defer mu.Unlock()
	decoded = map[string]image.Image{}
	decodedErr = map[string]error{}
	scaled = map[string][]scaledImage{}
}
