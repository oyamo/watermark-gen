package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// ReadImage Reads an image file and returns a *image.NRGBA struct
func ReadImage(path string) (image.Image, error) {
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
	return imgI, nil
}

// SaveImage Saves an image file into the secondary storage
func SaveImage(img image.Image, path string) error {
	var extension string

	// read raw file
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	// Parse file extension
	extension = filepath.Ext(path)
	if extension == "" {
		return fmt.Errorf("%s has to be of type png, jpeg or gif", path)
	}

	switch extension {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(file, img, nil)
		break
	case ".png":
		err = png.Encode(file, img)
		break
	case ".gif":
		err = gif.Encode(file, img, nil)
		break
	default:
		return fmt.Errorf("%s has to be of type png, jpeg or gif", path)
	}

	return err
}

func ResizeImage(img image.Image, height, width int) (image.Image, error) {
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

func AddWatermarkImage(mainImagePath, watermarkImagePath, outPath string, x, y, height, width int) error {
	// get mainImg image from the disk
	mainImg, err := ReadImage(mainImagePath)
	if err != nil {
		return err
	}

	// get the waterMarkImg image from the disk
	waterMarkImg, err := ReadImage(watermarkImagePath)
	if err != nil {
		return err
	}

	// resize image
	if waterMarkImg.Bounds().Dx() > width || waterMarkImg.Bounds().Dy() > height {
		waterMarkImg, err = ResizeImage(waterMarkImg, height, width)
		if err != nil {
			return err
		}
	}

	mainImageHeight := mainImg.Bounds().Dy()
	mainImageWidth := mainImg.Bounds().Dx()

	watermarkImageHeight := waterMarkImg.Bounds().Dy()
	watermarkImageWidth := waterMarkImg.Bounds().Dx()

	// Validate the dimensions
	if x < 0 || y < 0 {
		return errors.New("dimensions out of bounds")
	}

	if x > mainImageWidth || y > mainImageHeight {
		return errors.New("dimensions out of bounds")
	}

	var newImg *image.NRGBA

	// convert main image into *image.NRGBA
	if _, ok := mainImg.(*image.NRGBA); ok {
		newImg = mainImg.(*image.NRGBA)
	} else {
		newImg = image.NewNRGBA(image.Rect(0, 0, mainImageWidth, mainImageHeight))
		draw.Draw(newImg, newImg.Bounds(), mainImg, mainImg.Bounds().Min, draw.Src)
	}

	// Add waterMarkImg to the image
	for i := x; i < watermarkImageWidth+x; i++ {
		for j := y; j < watermarkImageHeight+y; j++ {
			waterMarkPixelColor := waterMarkImg.At(i-x, j-y)
			mainImagePixelColor := mainImg.At(i, j)
			blendedColor := Blend(waterMarkPixelColor, mainImagePixelColor)
			newImg.Set(i, j, blendedColor)
		}
	}

	err = SaveImage(newImg, outPath)
	if err != nil {
		return err
	}

	return nil
}

func ValidatePaths(path ...string) {
	for _, v := range path {
		if v == "" {
			fmt.Fprint(os.Stderr, "invalid usage\n")
			flag.Usage()
			os.Exit(1)
		}
	}
}
func main() {
	var mainImage, watermarkImage, outPath string
	var posX, posY, watermarkHeight, watermarkWidth int

	flag.StringVar(&mainImage, "m", "", "main image")
	flag.StringVar(&watermarkImage, "w", "", "watermark image")
	flag.StringVar(&outPath, "o", "", "out path")

	flag.IntVar(&posX, "x", 0, "x position on the main image")
	flag.IntVar(&posY, "y", 0, "y position on the main image")
	flag.IntVar(&watermarkHeight, "height", 0, "height of watermark")
	flag.IntVar(&watermarkWidth, "width", 0, "width of watermark")

	flag.Parse()

	ValidatePaths(mainImage, watermarkImage, outPath)

	// create watermark
	err := AddWatermarkImage(mainImage, watermarkImage, outPath, posX, posY, watermarkHeight, watermarkWidth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

}
