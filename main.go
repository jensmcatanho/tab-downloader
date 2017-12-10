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

func main() {
	url := fmt.Sprintf("https://www.ultimate-guitar.com/tabs/%v_guitar_pro_tabs.htm", os.Args[1])
	os.Mkdir(os.Args[1], os.ModePerm)

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

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
					downloadFile(name, tabID, tabURL)
				}
			})
		}
	})
}

func downloadFile(filename, id, referer string) {
	matches, _ := filepath.Glob(fmt.Sprintf("%v/%v", os.Args[1], filename) + ".*")
	for _, file := range matches {
		if _, err := os.Stat(file); err == nil {
			log.Println("File already exists. Fast-forwarding...")
			return
		}
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://tabs.ultimate-guitar.com/tab/download?id=%v", id), nil)
	req.Header.Add("Host", "tabs.ultimate-guitar.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0")
	req.Header.Add("Accept-Language", "pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", referer)
	req.Header.Add("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	splat := strings.Split(resp.Header.Get("Content-Disposition"), ".")
	extension := splat[len(splat)-1]
	out, err := os.Create(fmt.Sprintf("%v/%v.%v", os.Args[1], filename, extension))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
