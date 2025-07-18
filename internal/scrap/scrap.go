package scrap

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

func getFileName(url string) string {
	splits := strings.Split(url, "/")
	return splits[len(splits)-1]
}

func getFileType(filename string) string {
	splits := strings.Split(filename, ".")
	return splits[len(splits)-1]
}

func getEpisodeNumber(url string) string {
	splits := strings.Split(url, "-")
	for i, value := range splits {
		if value == "chapter" {
			return strings.TrimSuffix(splits[i+1], "/")
		}
	}
	return ""
}

func GetEpisode(url string) string {
	c := colly.NewCollector()
	episodeNumber := getEpisodeNumber(url)
	log.Println("Episode number:", episodeNumber)

	var wg sync.WaitGroup

	i := 0
	c.OnHTML("img", func(e *colly.HTMLElement) {
		imageUrl := e.Attr("src")
		wg.Add(1)
		log.Printf("Source found: %s\n", imageUrl)
		go func() {
			DownloadImage(imageUrl, episodeNumber, i)
			wg.Done()
		}()
		i++
	})

	os.Mkdir("./episodes/"+episodeNumber, 0700)
	c.Visit(url)

	wg.Wait()
	return episodeNumber
}

func GetHomePage(url string) []string {
	var chapters []string
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if strings.Contains(url, "chapter") {
			chapters = append(chapters, url)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
	removedIncomingChapters := chapters[4:]
	return removedIncomingChapters
}

func DownloadImage(url string, episodeNumber string, i int) {
	httpsPrefix := "https:"
	if !strings.HasPrefix(url, httpsPrefix) {
		url = httpsPrefix + url
	}
	response, err := http.Get(url)
	if err != nil {
		log.Println("Error while downloading image", err)
		return
	}
	if response.StatusCode != 200 {
		log.Printf("Couldn't fetch image. Status code: %d\n", response.StatusCode)
		return
	}

	defer response.Body.Close()

	filename := getFileName(url)
	filetype := getFileType(filename)

	filePath := fmt.Sprintf("./episodes/%s/%d.%s", episodeNumber, i, filetype)
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Couldn't create file :", filePath)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Image %s downloaded succesfully\n", url)
}
