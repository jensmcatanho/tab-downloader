package main

import (
	"fmt"
	"tab-downloader/representations"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	tabsQueue []representations.Tab
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Invalid number of arguments.")
	}

	processBands(os.Args)
}

func processBands(bands []string) {
	numFiles := 0
	processedChannel := make(chan bool, 500)
	doneChannel := make(chan bool, 500)

	go downloadWorker(processedChannel, doneChannel)
	for i := 1; i < len(os.Args); i++ {
		url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v", os.Args[i])
		os.MkdirAll(fmt.Sprintf("bands/%v", os.Args[i]), os.ModePerm)

		numPages, err := getNumberOfPages(url)
		if err != nil {
			log.Fatal(err)
		}

		for j := 1; j <= numPages; j++ {
			var numFilesAtPage int
			if j == 1 {
				numFilesAtPage, err = processPage(fmt.Sprint(url+"_guitar_pro_tabs.htm"), os.Args[i], processedChannel)
			} else {
				numFilesAtPage, err = processPage(fmt.Sprint(url+fmt.Sprintf("_guitar_pro_tabs%d.htm", j)), os.Args[i], processedChannel)
			}
	
			if err != nil {
				log.Fatal(err)
			}
	
			numFiles += numFilesAtPage
		}
	}

	for i := 0; i < numFiles; i++ {
		<-doneChannel
	}
}

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

func processPage(url, bandName string, processedChannel chan<- bool) (numFiles int, err error) {
	log.Printf("Processing page %v...", url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return
	}

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
					tabsQueue = append(tabsQueue, *representations.NewTab(tabName, bandName, tabID, tabURL))
					processedChannel <- true
					numFiles++
				}
			})
		}
	})

	return
}

func downloadWorker(processedChannel <-chan bool, doneChannel chan<- bool) {
	for true {
		<-processedChannel
		tab := tabsQueue[0]

		err := tab.Download()
		if err != nil {
			log.Printf("Couldn't download tab %v, error: %v", tab.Name, err)
		}

		tabsQueue = tabsQueue[1:]
		time.Sleep(1250 * time.Millisecond)
		doneChannel <- true
	}
}
