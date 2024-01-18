package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

func main() {
	var (
		input      string
		output     string
		width      int
		height     int
		cropX      int
		cropY      int
		cropWidth  int
		cropHeight int
		filter     string
		format     string
		batch      bool
	)

	flag.StringVar(&input, "input", "", "Input image file or directory (for batch processing).")
	flag.StringVar(&output, "output", "", "Output image file or directory (for batch processing).")
	flag.IntVar(&width, "width", 0, "Width to resize the image to.")
	flag.IntVar(&height, "height", 0, "Height to resize the image to.")
	flag.IntVar(&cropX, "crop-x", 0, "X coordinate for cropping the image.")
	flag.IntVar(&cropY, "crop-y", 0, "Y coordinate for cropping the image.")
	flag.IntVar(&cropWidth, "crop-width", 0, "Width of the cropped area.")
	flag.IntVar(&cropHeight, "crop-height", 0, "Height of the cropped area.")
	flag.StringVar(&filter, "filter", "", "Filter to apply to the image (blur, sharpen, etc.).")
	flag.StringVar(&format, "format", "", "Output image format (jpeg, png, etc.).")
	flag.BoolVar(&batch, "batch", false, "Enable batch processing.")

	flag.Parse()

	if input == "" || output == "" {
		fmt.Println("Input and output must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	if batch {
		err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && isImageFile(path) {
				processImage(path, output, width, height, cropX, cropY, cropWidth, cropHeight, filter, format)
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error processing batch:", err)
		}
	} else {
		processImage(input, output, width, height, cropX, cropY, cropWidth, cropHeight, filter, format)
	}
}

func processImage(input, output string, width, height, cropX, cropY, cropWidth, cropHeight int, filter, format string) {
	img := gocv.IMRead(input, gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error reading image:", input)
		return
	}
	defer img.Close()

	if width > 0 && height > 0 {
		gocv.Resize(img, &img, image.Pt(width, height), 0, 0, gocv.InterpolationDefault)
	}

	if cropWidth > 0 && cropHeight > 0 {
		cropRect := image.Rect(cropX, cropY, cropX+cropWidth, cropY+cropHeight)
		img = img.Region(cropRect)
	}

	if filter != "" {
		applyFilter(&img, filter)
	}

	if format != "" {
		output = changeExtension(output, format)
	}

	gocv.IMWrite(output, img)
}

func applyFilter(img *gocv.Mat, filter string) {
	switch filter {
	case "blur":
		gocv.GaussianBlur(*img, img, image.Pt(7, 7), 0, 0, gocv.BorderDefault)
	case "sharpen":
		kernel := gocv.NewMatWithSizeFromScalar(gocv.ScalarAll(-1), 3, 3, gocv.MatTypeCV64F)
		kernel.SetFloatAt(1, 1, 9)
		gocv.Filter2D(*img, img, -1, kernel, image.Pt(-1, -1), 0, gocv.BorderDefault)
		kernel.Close()
	}
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".bmp" || ext == ".gif"
}

func changeExtension(filename, newExt string) string {
	ext := filepath.Ext(filename)
	return filename[0:len(filename)-len(ext)] + "." + newExt
}
