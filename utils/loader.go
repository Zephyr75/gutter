package utils

import (
  "os"
  "image"
)

func GetImageFromFilePath(filePath string) (image.Image, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    image, _, err := image.Decode(f)
    return image, err
}



