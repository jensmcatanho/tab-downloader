package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Tab struct {
	name    string
	id      string
	referer string
}

var (
	tabs           []Tab
	tabsProcessed  chan bool
	tabsDone       chan bool
	pages          string
	pagesProcessed chan bool
	pagesDone      chan bool
)

/*
func main() {
	url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v_guitar_pro_tabs.htm", os.Args[1])
	tabsProcessed = make(chan bool, 100)
	tabsDone = make(chan bool, 100)
	pagesProcessed = make(chan bool, 100)
	pagesDone = make(chan bool, 100)
	numPages := 0
	numFiles := 0

}
func pagesGetter(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
}
*/

func main() {
	numPages := 0
	numFiles := 0
	url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v", os.Args[1])
	tabsProcessed = make(chan bool, 100)
	tabsDone = make(chan bool, 100)
	os.Mkdir(os.Args[1], os.ModePerm)

	doc, err := goquery.NewDocument(fmt.Sprint(url, "_guitar_pro_tabs.htm"))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("ys") {
			numPages++
		}
	})

	for i := 1; i <= numPages; i++ {
		var pageURL string
		if i == 1 {
			pageURL = fmt.Sprint(url, "_guitar_pro_tabs.htm")
		} else {
			pageURL = fmt.Sprint(url + fmt.Sprintf("_guitar_pro_tabs%d.htm", i))
		}

		log.Printf("Processing page %v...", pageURL)
		doc, err = goquery.NewDocument(pageURL)
		if err != nil {
			log.Fatal(err)
		}

		go downloadWorker()
		doc.Find("tr .tr__lg").Not(".tr__active").Each(func(i int, s *goquery.Selection) {
			if !s.Find("a").HasClass("song js-tp_link") {
				tab := s.Find("a")
				name := tab.Text()
				log.Printf("Processing %v tab...", name)
				tabURL, _ := tab.Attr("href")

				doc, err := goquery.NewDocument(tabURL)
				if err != nil {
					log.Fatal(err)
				}

				doc.Find("div").Each(func(i int, s *goquery.Selection) {
					if s.HasClass("textversbox") {
						tabID, _ := s.Find("input").Attr("value")
						tabs = append(tabs, Tab{name: name, id: tabID, referer: tabURL})
						tabsProcessed <- true
						numFiles++
					}
				})
			}
		})
	}

	for i := 0; i < numFiles; i++ {
		<-tabsDone
	}

}

func downloadWorker() {
	for true {
		<-tabsProcessed
		tab := tabs[0]
		fmt.Printf("%+v\n", tab)
		//downloadFile(tab)
		tabs = tabs[1:]
		//		time.Sleep(1250 * time.Millisecond)
		tabsDone <- true
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

	splat := strings.Split(resp.Header.Get("Content-Disposition"), ".")
	extension := splat[len(splat)-1]
	out, err := os.Create(fmt.Sprintf("%v/%v.%v", os.Args[1], tab.name, extension))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	log.Printf("Downloading %v file (#%v)...", tab.name, tab.id)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
