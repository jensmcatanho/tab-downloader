package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Tab struct {
	name    string
	id      string
	referer string
}

var (
	tabsQueue []Tab
	numTabs   int
)

func getNumberOfPages(url string) (numPages int, err error) {
	doc, err := goquery.NewDocument(fmt.Sprint(url, "_guitar_pro_tabs.htm"))
	if err != nil {
		log.Fatal(err)
		return
	}

	numPages++
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("ys") {
			numPages++
		}
	})

	return
}

func processPage(url string, processedChannel chan bool, doneChannel chan<- bool) (numFiles int, err error) {
	log.Printf("Processing page %v...", url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return
	}

	go downloadWorker(processedChannel, doneChannel)
	doc.Find("tr .tr__lg").Not(".tr__active").Each(func(i int, s *goquery.Selection) {
		if !s.Find("a").HasClass("song js-tp_link") {
			tab := s.Find("a")
			tabName := tab.Text()
			log.Printf("Processing %v tab...", tabName)
			tabURL, _ := tab.Attr("href")

			doc, err := goquery.NewDocument(tabURL)
			if err != nil {
				return
			}

			doc.Find("div").Each(func(i int, s *goquery.Selection) {
				if s.HasClass("textversbox") {
					tabID, _ := s.Find("input").Attr("value")
					tabsQueue = append(tabsQueue, Tab{name: tabName, id: tabID, referer: tabURL})
					processedChannel <- true
					numFiles++
				}
			})
		}
	})

	return
}

func main() {
	processedChannel := make(chan bool, 500)
	doneChannel := make(chan bool, 500)

	url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v", os.Args[1])
	os.Mkdir(os.Args[1], os.ModePerm)

	numPages, err := getNumberOfPages(url)
	if err != nil {
		log.Fatal(err)
	}

	numFiles := 0
	for i := 1; i <= numPages; i++ {
		var numFilesAtPage int
		if i == 1 {
			numFilesAtPage, err = processPage(fmt.Sprint(url+"_guitar_pro_tabs.htm"), processedChannel, doneChannel)
		} else {
			numFilesAtPage, err = processPage(fmt.Sprint(url+fmt.Sprintf("_guitar_pro_tabs%d.htm", i)), processedChannel, doneChannel)
		}

		if err != nil {
			log.Fatal(err)
		}

		numFiles += numFilesAtPage
	}

	for i := 0; i < numFiles; i++ {
		<-doneChannel
	}

	fmt.Println(numFiles)
}

func downloadWorker(processedChannel <-chan bool, doneChannel chan<- bool) {
	for true {
		<-processedChannel
		tab := tabsQueue[0]
		log.Printf("%+v", tab)
		//downloadFile(tab)
		tabsQueue = tabsQueue[1:]
		time.Sleep(1250 * time.Millisecond)
		doneChannel <- true
	}
}

func downloadFile(tab Tab) {
	matches, _ := filepath.Glob(fmt.Sprintf("%v/%v", os.Args[1], tab.name) + ".*")
	for _, file := range matches {
		if _, err := os.Stat(file); err == nil {
			log.Println("File already exists. Fast-forwarding...")
			return
		}
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://tabs.ultimate-guitar.com/tab/download?id=%v", tab.id), nil)
	req.Header.Add("Referer", tab.referer)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Printf("%+v\n", tab)

	splat := strings.Split(resp.Header.Get("Content-Disposition"), ".")
	extension := splat[len(splat)-1]
	out, err := os.Create(fmt.Sprintf("%v/%v.%v", os.Args[1], tab.name, extension))
	if err != nil {
		log.Printf("%v/%v.%v", os.Args[1], tab.name, extension)
		log.Fatal(err)
	}
	defer out.Close()

	log.Printf("Downloading %v file (#%v)...", tab.name, tab.id)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
