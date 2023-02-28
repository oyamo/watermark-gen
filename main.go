package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
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

func ResizeImage(img *image.NRGBA, height, width int) (*image.NRGBA, error) {
	if img == nil {
		return nil, errors.New("image is nil")
	}

	currentBounds := img.Bounds()
	newBounds := image.Rect(0, 0, width, height)
	newImage := image.NewNRGBA(newBounds)
	for i := 0; i < newBounds.Dx(); i++ {
		for j := 0; j < newBounds.Dy(); j++ {
			atX := int(float64(i) * float64(currentBounds.Dx()) / float64(newBounds.Dx()))
			atY := int(float64(j) * float64(currentBounds.Dy()) / float64(newBounds.Dy()))
			colorAt := img.At(atX, atY)
			R, G, B, A := colorAt.RGBA()
			colorAtRGBA := color.NRGBA{R: uint8(R), G: uint8(G), B: uint8(B), A: uint8(A)}
			newImage.SetNRGBA(i, j, colorAtRGBA)
		}
	}

	return newImage, nil
}

func Blend(watermark color.Color, main color.Color) color.Color {
	wr, wg, wb, wa := watermark.RGBA()
	mr, mg, mb, ma := main.RGBA()

	// If the watermark pixel is fully transparent, return the main pixel.
	if wa == 0 {
		return main
	}

	// If the watermark pixel is fully opaque, return the watermark pixel.
	if wa == 0xffff {
		return watermark
	}

	// Calculate the blended color using the alpha values of the two pixels.
	alpha := float64(wa) / float64(0xffff)
	r := uint16(float64(wr)*alpha + float64(mr)*(1-alpha))
	g := uint16(float64(wg)*alpha + float64(mg)*(1-alpha))
	b := uint16(float64(wb)*alpha + float64(mb)*(1-alpha))
	a := uint16(math.Max(float64(wa), float64(ma)))

	return color.RGBA64{R: r, G: g, B: b, A: a}
}

func AddWatermarkImage(mainImagePath, watermarkImagePath string, x, y int) (image.Image, error) {
	main, err := ReadImage(mainImagePath)
	if err != nil {
		return nil, err
	}

	watermark, err := ReadImage(watermarkImagePath)
	if err != nil {
		return nil, err
	}

	mainImageHeight := main.Bounds().Dy()
	mainImageWidth := main.Bounds().Dx()

	watermarkImageHeight := watermark.Bounds().Dy()
	watermarkImageWidth := watermark.Bounds().Dx()

	if x < 0 || y < 0 {
		return nil, errors.New("dimensions out of bounds")
	}

	if x > mainImageWidth || y > mainImageHeight {
		return nil, errors.New("dimensions out of bounds")
	}

	for i := x; i < watermarkImageWidth+x; i++ {
		for j := y; j < watermarkImageHeight+y; j++ {
			waterMarkPixelColor := watermark.At(i-x, j-y)
			mainImagePixelColor := main.At(i, j)

			blendedColor := Blend(waterMarkPixelColor, mainImagePixelColor)
			main.Set(i, j, blendedColor)
		}
	}

	return main, nil
}
