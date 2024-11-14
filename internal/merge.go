package internal

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/go-pdf/fpdf"
)

const (
	DPI                = 96
	MM_IN_INCH float64 = 25.4
	A4_HEIGHT          = 297
	A4_WIDTH           = 210
	MAX_WIDTH  float64 = 800
	MAX_HEIGHT float64 = 500
)

func getSize(imgPath string) image.Config {
	file, err := os.Open(imgPath)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.DecodeConfig(file)
	return img
}

func pixelToMM(val float64) float64 {
	return float64(val) * MM_IN_INCH / DPI
}

func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

func getSizeScaled(width, height float64) (float64, float64) {
	widthScale := MAX_WIDTH / width
	heightScale := MAX_HEIGHT / height
	scale := min(widthScale, heightScale)
	return pixelToMM(scale * width), pixelToMM(scale * height)
}

func Merge() {
	folder := "1131"
	images := 13
	var opt fpdf.ImageOptions
	opt.ReadDpi = true
	opt.ImageType = "jpg"
	pdf := fpdf.New("P", "mm", "A4", "")
	for i := 0; i <= images; i++ {
		imagePath := fmt.Sprintf("./%s/%d.%s", folder, i, opt.ImageType)
		imageConfig := getSize(imagePath)
		width, height := getSizeScaled(float64(imageConfig.Width), float64(imageConfig.Height))
		log.Printf("Width: %f Heigh: %f for image: %s\n", width, height, imagePath)
		pdf.AddPage()
		pdf.ImageOptions(imagePath, 0, 0, width, height, false, opt, 0, "")
	}
	err := pdf.OutputFileAndClose("example.pdf")
	if err != nil {
		log.Println("error: ", err)
	}
}
