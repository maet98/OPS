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
		go DownloadImage(src, episodeNumber, i, &wg)
		i++
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	os.Mkdir("./episodes/"+episodeNumber, 0700)
	c.Visit(url)

	wg.Wait()
	return episodeNumber
}

func GetHomePage(url string) []string {
	var answer []string
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if strings.Contains(url, "chapter") {
			log.Printf("Found new episode %s -> %s", e.Text, url)
			answer = append(answer, url)
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
	return answer
}

func DownloadImage(url string, episodeNumber string, i int, wg *sync.WaitGroup) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	filename := getFileName(url)
	filetype := getFileType(filename)

	file, err := os.Create(fmt.Sprintf("./episodes/%s/%d.%s", episodeNumber, i, filetype))
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
