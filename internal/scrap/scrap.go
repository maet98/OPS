package scrap

import (
	"fmt"
	"io"
	"log"
	"maet98/scrapper/internal/config"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
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
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String("/usr/bin/chromium"),
	})

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	episodeNumber := getEpisodeNumber(url)
	config := config.GetConfig()
	log.Println("Episode number:", episodeNumber)
	os.MkdirAll(config.BaseFolder+"/"+episodeNumber, 0700)

	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	entries, err := page.Locator("img").All()
	if err != nil {
		log.Fatalf("could not get images: %v", err)
	}

	var wg sync.WaitGroup
	for i, entry := range entries {
		imageUrl, err := entry.GetAttribute("src")
		if err != nil || imageUrl == "" {
			continue
		}
		if !strings.Contains(imageUrl, "chapter") {
			continue
		}

		log.Printf("Source found: %s\n", imageUrl)
		wg.Add(1)
		go func(url string, index int) {
			defer wg.Done()
			DownloadImage(url, episodeNumber, index)
		}(imageUrl, i)
	}

	wg.Wait()
	return episodeNumber
}

func GetHomePage(url string) []string {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	log.Println("Visiting", url)
	if _, err = page.Goto(url); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	locators, err := page.Locator("a[href*='chapter']").All()
	if err != nil {
		log.Fatal(err)
	}

	var chapters []string
	for _, link := range locators {
		href, _ := link.GetAttribute("href")
		if href != "" {
			chapters = append(chapters, href)
		}
	}

	if len(chapters) > 4 {
		return chapters[4:]
	}
	return chapters
}

func DownloadImage(url string, episodeNumber string, i int) {
	if strings.HasPrefix(url, "//") {
		url = "https:" + url
	} else if !strings.HasPrefix(url, "http") {
		url = "https:" + url
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	response, err := client.Do(req)
	if err != nil {
		log.Println("Error while downloading image", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("Couldn't fetch image. Status code: %d\n", response.StatusCode)
		return
	}

	filename := getFileName(url)
	filetype := getFileType(filename)
	if strings.Contains(filetype, "?") {
		filetype = strings.Split(filetype, "?")[0]
	}

	filePath := fmt.Sprintf(config.GetConfig().BaseFolder+"/%s/%d.%s", episodeNumber, i, filetype)
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

	log.Printf("Image %s downloaded successfully\n", url)
}
