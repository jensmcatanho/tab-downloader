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
	numFiles  int
	tabs      []Tab
	processed chan bool
	done      chan bool
)

func main() {
	url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v_guitar_pro_tabs.htm", os.Args[1])
	processed = make(chan bool, 100)
	done = make(chan bool, 100)
	os.Mkdir(os.Args[1], os.ModePerm)

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	numFiles = 0
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
					processed <- true
					numFiles++
					// downloadFile(name, tabID, tabURL)
				}
			})
		}
	})

	for i := 0; i < numFiles; i++ {
		<-done
	}
}

func downloadWorker() {
	for true {
		<-processed
		tab := tabs[0]
		downloadFile(tab)
		tabs = tabs[1:]
		time.Sleep(1250 * time.Millisecond)
		done <- true
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
