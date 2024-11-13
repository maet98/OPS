package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/unidoc/unipdf/v3/creator"
)

func getFileName(url string) string {
	splits := strings.Split(url, "/")
	return splits[len(splits)-1]
}

func getFileType(filename string) string {
	splits := strings.Split(filename, ".")
	return splits[len(splits)-1]
}

func mergeToPdf(folder string) error {
	c := creator.New()
	dir := "./" + folder
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d files to merge\n", len(files))

	for _, file := range files {
		imgPath := "./" + folder + "/" + file.Name()
		img, err := c.NewImageFromFile(imgPath)
		if err != nil {
			log.Println("Found error when opening image", err)
			return err
		}
		pageWidth := 612.0
		img.ScaleToWidth(pageWidth)

		pageHeight := pageWidth * img.Height() / img.Width()
		c.SetPageSize(creator.PageSize{pageWidth, pageHeight})
		c.NewPage()
		img.SetPos(0, 0)
		_ = c.Draw(img)
	}

	err = c.WriteToFile("./test")
	return err
}

func downloadImage(url string, episodeNumber string, i int, wg *sync.WaitGroup) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	filename := getFileName(url)
	filetype := getFileType(filename)

	file, err := os.Create(fmt.Sprintf("./%s/%d.%s", episodeNumber, i, filetype))
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Imaged downloaded succesfully")
	wg.Done()
}

func getEpisodeNumber(url string) string {
	splits := strings.Split(url, "-")
	return splits[len(splits)-1]
}

func scrap(url string) {
	c := colly.NewCollector()
	// Instantiate default collector
	episodeNumber := getEpisodeNumber(url)
	log.Println("Episode number:", episodeNumber)

	var wg sync.WaitGroup

	i := 0
	// On every a element which has href attribute call callback
	c.OnHTML("img", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		// Print link
		wg.Add(1)
		fmt.Printf("Source found: %s\n", src)
		go downloadImage(src, episodeNumber, i, &wg)
		i++
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	os.Mkdir(episodeNumber, 0700)
	c.Visit(url)

	wg.Wait()
}

func main() {
	url := "https://w43.1piecemanga.com/manga/one-piece-chapter-1119/"
	episodeNumber := getEpisodeNumber(url)
	err := mergeToPdf(episodeNumber)
	log.Println("Error", err)
}