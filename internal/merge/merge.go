package merge

import (
	"fmt"
	"image"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

func MergeToPdf(episodeNumber string) {
	images := getImages(episodeNumber)
	var opt fpdf.ImageOptions
	opt.ReadDpi = true
	pdf := fpdf.New("P", "mm", "A4", "")
	for _, imageName := range images {
		imageType := getImageType(imageName)
		if !isSupported(imageType) {
			continue
		}

		opt.ImageType = imageType
		imagePath := fmt.Sprintf("./episodes/%s/%s", episodeNumber, imageName)

		imageConfig := getSize(imagePath)
		width, height := getSizeScaled(float64(imageConfig.Width), float64(imageConfig.Height), pdf)

		if width > height {
			pdf.AddPageFormat("H", pdf.GetPageSizeStr("A4"))
		} else {
			pdf.AddPage()
		}
		pdf.ImageOptions(imagePath, 0, 0, width, height, false, opt, 0, "")
	}

	chapterName := fmt.Sprintf("./episodes/chapter-%s.pdf", episodeNumber)
	err := pdf.OutputFileAndClose(chapterName)
	if err != nil {
		log.Println("error: ", err)
	}
}

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
	var scale float64
	if height > width {
		widthScale := size.Wd / width
		heightScale := size.Ht / height
		scale = min(widthScale, heightScale)
	} else {
		widthScale := size.Ht / width
		heightScale := size.Wd / height
		scale = min(widthScale, heightScale)
	}

	return scale * width, scale * height
}

func getImages(episodeNumber string) []string {
	f, err := os.OpenFile("./episodes/"+episodeNumber, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	files, _ := f.Readdirnames(0)
	slices.SortFunc(files, func(a, b string) int {
		aRaw := strings.Split(a, ".")[0]
		aNumber, err := strconv.Atoi(aRaw)
		if err != nil {
			log.Fatal(err)
		}

		bRaw := strings.Split(b, ".")[0]
		bNumber, err := strconv.Atoi(bRaw)
		if err != nil {
			log.Fatal(err)
		}
		return aNumber - bNumber
	})
	return files
}

func getImageType(imageName string) string {
	splits := strings.Split(imageName, ".")
	return splits[len(splits)-1]
}

func isSupported(imageType string) bool {
	supportedImageType := []string{
		"jpg",
		"jpeg",
		"png",
		"gif",
	}
	for _, sup := range supportedImageType {
		if sup == imageType {
			return true
		}
	}

	return false
}
