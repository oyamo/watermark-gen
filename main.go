package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// ReadImage Reads an image file and returns a *image.NRGBA struct
func ReadImage(path string) (*image.NRGBA, error) {
	var extension string
	var imgI image.Image // image.Image interface

	// read raw file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Parse file extension
	extension = filepath.Ext(path)
	if extension == "" {
		return nil, fmt.Errorf("%s has to be of type png, jpeg or gif", path)
	}

	switch extension {
	case ".jpg", ".jpeg":
		imgI, err = jpeg.Decode(file)
		break
	case ".png":
		imgI, err = png.Decode(file)
		break
	case ".gif":
		imgI, err = gif.Decode(file)
		break
	default:
		return nil, fmt.Errorf("%s has to be of type png, jpeg or gif", path)
	}

	if err != nil {
		return nil, err
	}

	// Cast the interface to struct
	return imgI.(*image.NRGBA), nil
}
