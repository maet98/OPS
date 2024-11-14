package internal

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/go-pdf/fpdf"
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

func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

func getSizeScaled(width, height float64, pdf *fpdf.Fpdf) (float64, float64) {
	size := pdf.GetPageSizeStr("A4")

	log.Printf("Width: %f Heigh: %f \n", size.Wd, size.Ht)
	widthScale := size.Wd / width
	heightScale := size.Ht / height
	scale := min(widthScale, heightScale)
	return scale * width, scale * height
}

func Merge() {
	folder := "1119"
	images := 11
	var opt fpdf.ImageOptions
	opt.ReadDpi = true
	opt.ImageType = "png"
	pdf := fpdf.New("P", "mm", "A4", "")
	for i := 0; i <= images; i++ {
		imagePath := fmt.Sprintf("./%s/%d.%s", folder, i, opt.ImageType)
		imageConfig := getSize(imagePath)
		width, height := getSizeScaled(float64(imageConfig.Width), float64(imageConfig.Height), pdf)
		log.Printf("Width: %f Heigh: %f for image: %s\n", width, height, imagePath)

		if width > height {
			pdf.AddPageFormat("H", pdf.GetPageSizeStr("A4"))
		} else {
			pdf.AddPage()
		}
		pdf.ImageOptions(imagePath, 0, 0, width, height, false, opt, 0, "")

	}
	err := pdf.OutputFileAndClose("example.pdf")
	if err != nil {
		log.Println("error: ", err)
	}
}
