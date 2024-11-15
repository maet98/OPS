package internal

import (
	"fmt"
	"image"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/go-pdf/fpdf"
)

func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

func getSize(imgPath string) image.Config {
	file, err := os.Open(imgPath)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.DecodeConfig(file)
	return img
}

func getSizeScaled(width, height float64, pdf *fpdf.Fpdf) (float64, float64) {
	size := pdf.GetPageSizeStr("A4")

	log.Printf("Width: %f Heigh: %f \n", size.Wd, size.Ht)
	widthScale := size.Wd / width
	heightScale := size.Ht / height
	scale := min(widthScale, heightScale)
	return scale * width, scale * height
}

func GetImages(folder string) []string {
	var images []string
	f, err := os.OpenFile("./episodes/"+folder, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	files, _ := f.Readdirnames(0)
	slices.Sort(files)
	log.Println(files)
	return images
}

func getImageType(imageName string) string {
	splits := strings.Split(imageName, ".")
	return splits[len(splits)-1]
}

func Merge(folder string) {
	images := GetImages(folder)
	var opt fpdf.ImageOptions
	opt.ReadDpi = true
	pdf := fpdf.New("P", "mm", "A4", "")
	for _, imageName := range images {
		opt.ImageType = getImageType(imageName)
		imagePath := fmt.Sprintf("./episodes/%s%s", folder, imageName)
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
